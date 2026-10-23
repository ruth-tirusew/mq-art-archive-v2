package postgres

import "strings"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// containsPattern builds a "%text%" pattern for ILIKE with the user's own % and _ treated as plain
// characters. PostgreSQL's default LIKE escape is a backslash, so no ESCAPE clause is needed.
// Without this, a search for "%" matches everything and "_" matches any single character.
func containsPattern(text string) string {
	return "%" + likeEscaper.Replace(text) + "%"
}
