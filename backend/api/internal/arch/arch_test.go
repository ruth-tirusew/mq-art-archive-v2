package arch

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulePrefix = "github.com/mq/api/"

var repositoryFileToContext = map[string]string{
	"article_repository.go":     "content",
	"art_post_repository.go":    "art",
	"profile_repository.go":     "profile",
	"institution_repository.go": "institution",
	"onboarding_repository.go":  "onboarding",
	"user_repository.go":        "identity",
	"oauth_account_repository.go": "identity",
	"event_repository.go":         "events",
	"event_location_repository.go": "events",
}

var usecaseDirToContext = map[string]string{
	"art":         "art",
	"content":     "content",
	"identity":    "identity",
	"auth":        "identity",
	"institution": "institution",
	"onboarding":  "onboarding",
	"profile":     "profile",
	"events":      "events",
}

type layer int

const (
	layerUnknown layer = iota
	layerDomain
	layerAppErrors
	layerPortInbound
	layerPortOutbound
	layerUsecase
	layerAdapterDriving
	layerAdapterDriven
	layerConfig
	layerTestutil
)

func TestArchitecture(t *testing.T) {
	root, err := moduleRoot()
	if err != nil {
		t.Fatalf("module root: %v", err)
	}

	var violations []string

	err = filepath.Walk(filepath.Join(root, "internal"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		imports, err := fileImports(path)
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(root, path)
		pkgLayer := classifyPackage(rel)

		for _, imp := range imports {
			if !strings.HasPrefix(imp, modulePrefix) {
				continue
			}
			if v := checkImport(rel, path, pkgLayer, imp); v != "" {
				violations = append(violations, v)
			}
		}

		if v := checkPersistenceVerticalBoundary(rel, path, imports); v != "" {
			violations = append(violations, v)
		}
		if v := checkUsecaseVerticalBoundary(rel, imports); v != "" {
			violations = append(violations, v)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("architecture violations (%d):\n%s", len(violations), strings.Join(violations, "\n"))
	}
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func fileImports(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	var imports []string
	for _, imp := range f.Imports {
		imports = append(imports, strings.Trim(imp.Path.Value, `"`))
	}
	return imports, nil
}

func classifyPackage(rel string) layer {
	p := filepath.ToSlash(rel)
	switch {
	case strings.HasPrefix(p, "internal/domain/apperrors/"):
		return layerAppErrors
	case strings.HasPrefix(p, "internal/domain/"):
		return layerDomain
	case strings.HasPrefix(p, "internal/port/inbound/"):
		return layerPortInbound
	case strings.HasPrefix(p, "internal/port/outbound/"):
		return layerPortOutbound
	case strings.HasPrefix(p, "internal/usecase/"):
		return layerUsecase
	case strings.HasPrefix(p, "internal/adapter/driving/"):
		return layerAdapterDriving
	case strings.HasPrefix(p, "internal/adapter/driven/"):
		return layerAdapterDriven
	case strings.HasPrefix(p, "internal/testutil/"):
		return layerTestutil
	default:
		return layerUnknown
	}
}

func importLayer(imp string) layer {
	suffix := strings.TrimPrefix(imp, modulePrefix)
	switch {
	case strings.HasPrefix(suffix, "internal/domain/apperrors"):
		return layerAppErrors
	case strings.HasPrefix(suffix, "internal/domain/"):
		return layerDomain
	case strings.HasPrefix(suffix, "internal/port/inbound"):
		return layerPortInbound
	case strings.HasPrefix(suffix, "internal/port/outbound"):
		return layerPortOutbound
	case strings.HasPrefix(suffix, "internal/usecase/"):
		return layerUsecase
	case strings.HasPrefix(suffix, "internal/adapter/driving/"):
		return layerAdapterDriving
	case strings.HasPrefix(suffix, "internal/adapter/driven/"):
		return layerAdapterDriven
	case strings.HasPrefix(suffix, "internal/testutil/"):
		return layerTestutil
	case strings.HasPrefix(suffix, "config"):
		return layerConfig
	default:
		return layerUnknown
	}
}

func checkImport(rel, path string, pkgLayer layer, imp string) string {
	target := importLayer(imp)

	switch pkgLayer {
	case layerDomain:
		if target != layerUnknown {
			return fmt.Sprintf("%s: domain must not import internal package %s", rel, imp)
		}
	case layerAppErrors:
		if target != layerUnknown {
			return fmt.Sprintf("%s: apperrors must not import internal package %s", rel, imp)
		}
	case layerPortInbound, layerPortOutbound:
		if target == layerAdapterDriving || target == layerAdapterDriven || target == layerUsecase {
			return fmt.Sprintf("%s: port must not import %s", rel, imp)
		}
	case layerUsecase:
		if target == layerAdapterDriving || target == layerAdapterDriven {
			return fmt.Sprintf("%s: usecase must not import adapter %s", rel, imp)
		}
	case layerAdapterDriving:
		if target == layerAdapterDriven {
			return fmt.Sprintf("%s: driving adapter must not import driven adapter %s", rel, imp)
		}
		if target == layerUsecase {
			return fmt.Sprintf("%s: driving adapter must not import usecase %s (use port/inbound)", rel, imp)
		}
	case layerAdapterDriven:
		if strings.Contains(rel, "persistence/postgres") {
			if target == layerUsecase {
				return fmt.Sprintf("%s: persistence must not import usecase %s", rel, imp)
			}
			if target == layerPortInbound {
				return fmt.Sprintf("%s: persistence must not import port/inbound %s", rel, imp)
			}
			if target == layerAdapterDriving {
				return fmt.Sprintf("%s: persistence must not import driving adapter %s", rel, imp)
			}
		}
	}
	_ = path
	return ""
}

func checkPersistenceVerticalBoundary(rel, path string, imports []string) string {
	if !strings.Contains(rel, "internal/adapter/driven/persistence/postgres/") {
		return ""
	}
	file := filepath.Base(path)
	expected, ok := repositoryFileToContext[file]
	if !ok {
		return ""
	}

	for _, imp := range imports {
		if !strings.HasPrefix(imp, modulePrefix+"internal/domain/") {
			continue
		}
		if strings.HasPrefix(imp, modulePrefix+"internal/domain/apperrors") {
			continue
		}
		ctx := domainContextFromImport(imp)
		if ctx != expected {
			return fmt.Sprintf("%s: persistence may only import domain/%s, not domain/%s", rel, expected, ctx)
		}
	}
	return ""
}

func checkUsecaseVerticalBoundary(rel string, imports []string) string {
	if !strings.HasPrefix(rel, "internal/usecase/") {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) < 3 {
		return ""
	}
	usecaseCtx, ok := usecaseDirToContext[parts[2]]
	if !ok {
		return ""
	}

	for _, imp := range imports {
		if !strings.HasPrefix(imp, modulePrefix+"internal/domain/") {
			continue
		}
		if strings.HasPrefix(imp, modulePrefix+"internal/domain/apperrors") {
			continue
		}
		ctx := domainContextFromImport(imp)
		if ctx != usecaseCtx {
			return fmt.Sprintf("%s: usecase/%s may only import domain/%s, not domain/%s", rel, parts[2], usecaseCtx, ctx)
		}
	}
	return ""
}

func domainContextFromImport(imp string) string {
	suffix := strings.TrimPrefix(imp, modulePrefix+"internal/domain/")
	return strings.Split(suffix, "/")[0]
}
