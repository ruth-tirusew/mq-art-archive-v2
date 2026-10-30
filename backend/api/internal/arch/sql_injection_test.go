package arch

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// SQL injection is prevented by always sending user input as bind parameters ($1, $2, ...), never
// by building the query text from it. This test is a tripwire for the two ways that rule is
// usually broken in Go, applied to every non-test file in the Postgres adapter:
//
//  1. fmt.Sprintf with a SQL-looking format string and any verb other than %d (%s, %v, %q ... all
//     splice text into the query; %d is allowed for building "$1, $2" placeholder numbers).
//  2. Concatenating a function parameter (or other caller-controlled string) into the query text
//     passed to Query/QueryRow/Exec.
//
// Fragments built from constants (column lists, "WHERE role=$1") are fine.

var sqlKeyword = regexp.MustCompile(`(?i)\b(select|insert|update|delete|where|order by|limit|offset|from|set)\b`)
var nonIntVerb = regexp.MustCompile(`%[^%d]`)

var sqlExecMethods = map[string]bool{"Query": true, "QueryRow": true, "Exec": true}

// sqlHelpers take SQL text (or a fragment of it) as a parameter and forward it to the driver. That
// is allowed only because every call to them is also checked: the SQL argument (at the index given
// here) must be built from literals and constants, never from the caller's own parameters.
var sqlHelpers = map[string]int{"getOne": 1, "list": 1, "queryEvents": 1}

func TestPostgresAdapterUsesBindParameters(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatalf("module root: %v", err)
	}
	dir := filepath.Join(root, "internal", "adapter", "driven", "persistence", "postgres")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	var violations []string
	checked := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		violations = append(violations, checkSQLSafety(name, string(src))...)
		checked++
	}
	if checked < 10 {
		t.Fatalf("only scanned %d files; is the path right?", checked)
	}
	if len(violations) > 0 {
		t.Fatalf("SQL must use bind parameters, not string building (%d):\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

// TestSQLGuardCatchesInjectionPatterns proves the guard above can actually fail.
func TestSQLGuardCatchesInjectionPatterns(t *testing.T) {
	bad := map[string]string{
		"sprintf %s": `package p
import "fmt"
func f(r R, name string) { r.Query(ctx, fmt.Sprintf("SELECT * FROM t WHERE name = '%s'", name)) }`,
		"concat param": `package p
func f(r R, name string) { r.Exec(ctx, "DELETE FROM t WHERE name = '" + name + "'") }`,
		"bare param": `package p
func f(r R, q string) { r.QueryRow(ctx, q) }`,
		"method call result": `package p
func f(r R, id I) { r.Query(ctx, "SELECT 1 FROM t WHERE id = '"+id.String()+"'") }`,
		"field of param": `package p
func f(r R, in In) { r.Exec(ctx, "DELETE FROM t WHERE n = '"+in.Name+"'") }`,
		"tainted local": `package p
func f(r R, name string) { q := "SELECT 1 FROM t WHERE n = '" + name + "'"; r.Query(ctx, q) }`,
		"tainted local appended": `package p
func f(r R, name string) { q := "SELECT 1 FROM t WHERE"; q += " n = '" + name + "'"; r.Query(ctx, q) }`,
		"local from method call": `package p
func f(r R, id I) { x := id.String(); r.Query(ctx, "SELECT 1 FROM t WHERE id = '"+x+"'") }`,
		"helper with non-literal": `package p
func (r *X) getOne(ctx C, where string) { r.pool.QueryRow(ctx, "SELECT 1 FROM t WHERE "+where) }
func (r *X) g(name string) { r.getOne(ctx, name) }`,
		"list helper with concat param": `package p
func (r *X) list(ctx C, query string) { r.pool.Query(ctx, query) }
func (r *X) g(name string) { r.list(ctx, "SELECT 1 WHERE n = '"+name+"'") }`,
	}
	for name, src := range bad {
		if v := checkSQLSafety("bad.go", src); len(v) == 0 {
			t.Errorf("%s: expected a violation, got none", name)
		}
	}

	good := map[string]string{
		"constant concat": `package p
const cols = "id, name"
func f(r R) { r.Query(ctx, "SELECT "+cols+" FROM t WHERE id = $1", 1) }`,
		"sprintf %d placeholders": `package p
import "fmt"
func f(r R, n int) { r.Query(ctx, "SELECT 1 FROM t"+fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)) }`,
		"helper with literal": `package p
func (r *X) getOne(ctx C, where string) { r.pool.QueryRow(ctx, "SELECT 1 FROM t WHERE "+where) }
func (r *X) g() { r.getOne(ctx, "id = $1", 1) }`,
		"local built from constants": `package p
func f(r R, byRole bool) { where := ""; if byRole { where = " WHERE role = $1" }; r.Query(ctx, "SELECT 1 FROM t"+where) }`,
		"helper with constant columns": `package p
const cols = "id"
func (r *X) list(ctx C, query string) { r.pool.Query(ctx, query) }
func (r *X) g() { r.list(ctx, "SELECT "+cols+" FROM t WHERE a=$1", 1) }`,
	}
	for name, src := range good {
		if v := checkSQLSafety("good.go", src); len(v) != 0 {
			t.Errorf("%s: unexpected violations: %v", name, v)
		}
	}
}

func checkSQLSafety(file, src string) []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, 0)
	if err != nil {
		return []string{file + ": parse error: " + err.Error()}
	}
	var out []string
	report := func(n ast.Node, msg string) {
		out = append(out, file+":"+itoa(fset.Position(n.Pos()).Line)+": "+msg)
	}

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		params := map[string]bool{}
		if fn.Type.Params != nil {
			for _, p := range fn.Type.Params.List {
				for _, n := range p.Names {
					params[n.Name] = true
				}
			}
		}
		_, allowParams := sqlHelpers[fn.Name.Name]
		// locals that were assigned a non-constant string; using them as SQL is flagged like a parameter.
		locals := map[string]bool{}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if as, ok := n.(*ast.AssignStmt); ok && len(as.Lhs) == len(as.Rhs) {
				for i, lhs := range as.Lhs {
					id, ok := lhs.(*ast.Ident)
					if !ok {
						continue
					}
					// `x := <tainted>`, `x = <tainted>` and `x += <tainted>` all make x tainted.
					if tainted(as.Rhs[i], params, locals, false) != "" {
						locals[id.Name] = true
					}
				}
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			// Rule 1: Sprintf with SQL text may only format integers.
			if isSelector(call.Fun, "fmt", "Sprintf") && len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING &&
					sqlKeyword.MatchString(lit.Value) && nonIntVerb.MatchString(lit.Value) {
					report(call, "fmt.Sprintf builds SQL with a non-%d verb; use bind parameters")
				}
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			// Callers of SQL helpers must pass SQL built only from literals and constants.
			if idx, isHelper := sqlHelpers[sel.Sel.Name]; isHelper && len(call.Args) > idx {
				if bad := tainted(call.Args[idx], params, locals, false); bad != "" {
					report(call, sel.Sel.Name+" called with SQL that includes "+bad)
				}
			}
			// Rule 2: the query text of Query/QueryRow/Exec must not contain caller-controlled strings.
			if sqlExecMethods[sel.Sel.Name] && len(call.Args) >= 2 {
				if bad := tainted(call.Args[1], params, locals, allowParams); bad != "" {
					report(call, "query text includes "+bad+"; pass it as a bind parameter")
				}
			}
			return true
		})
	}
	return out
}

// tainted returns a description of the first operand of a SQL string expression that is not a
// string literal, a constant-style identifier, or a fmt.Sprintf (checked separately for verbs).
// Function parameters are flagged unless allowParams is set (for the SQL helpers). Any call
// (id.String(), strings.ToLower(x)), field access or index is flagged, because it can carry a
// caller-controlled value into the query text. Returns "" when the expression is safe.
func tainted(e ast.Expr, params, locals map[string]bool, allowParams bool) string {
	switch v := e.(type) {
	case *ast.BasicLit:
		return ""
	case *ast.ParenExpr:
		return tainted(v.X, params, locals, allowParams)
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "a non-concatenation expression"
		}
		if s := tainted(v.X, params, locals, allowParams); s != "" {
			return s
		}
		return tainted(v.Y, params, locals, allowParams)
	case *ast.Ident:
		if params[v.Name] && !allowParams {
			return "parameter " + v.Name
		}
		if locals[v.Name] {
			return "local variable " + v.Name + " built from a non-constant value"
		}
		return ""
	case *ast.CallExpr:
		if isSelector(v.Fun, "fmt", "Sprintf") {
			return ""
		}
		return "the result of a function call"
	default:
		return "a computed value"
	}
}

func isSelector(e ast.Expr, pkg, name string) bool {
	s, ok := e.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != name {
		return false
	}
	id, ok := s.X.(*ast.Ident)
	return ok && id.Name == pkg
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
