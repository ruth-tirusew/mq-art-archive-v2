package profile

import (
	"strings"
)

func slugify(value string) string {
	s := strings.ToLower(strings.TrimSpace(value))
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
