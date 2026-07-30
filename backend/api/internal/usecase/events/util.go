package events

import (
	"net/url"
	"path"
	"regexp"
	"strings"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "event"
	}
	if len(s) > 80 {
		s = s[:80]
		s = strings.Trim(s, "-")
	}
	return s
}

// uniqueSlug builds a title slug plus a short source-URL suffix so scraped
// posts with identical titles (common on Telegram) do not collide.
func uniqueSlug(title, sourceURL string) string {
	base := slugify(title)
	suffix := sourceSlugSuffix(sourceURL)
	if suffix == "" || suffix == base || strings.HasSuffix(base, "-"+suffix) {
		return base
	}
	maxBase := 100 - 1 - len(suffix)
	if maxBase < 1 {
		return suffix
	}
	if len(base) > maxBase {
		base = strings.Trim(base[:maxBase], "-")
	}
	if base == "" {
		return suffix
	}
	return base + "-" + suffix
}

func sourceSlugSuffix(sourceURL string) string {
	sourceURL = strings.TrimSpace(sourceURL)
	if sourceURL == "" {
		return ""
	}
	if u, err := url.Parse(sourceURL); err == nil {
		if seg := path.Base(strings.TrimSuffix(u.Path, "/")); seg != "" && seg != "." && seg != "/" {
			return slugify(seg)
		}
	}
	parts := strings.Split(strings.Trim(sourceURL, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return slugify(parts[len(parts)-1])
}
