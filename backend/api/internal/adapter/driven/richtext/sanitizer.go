// Package richtext sanitises the small subset of HTML that rich text (questionnaire answers, and
// any other field written with the web editor) may contain.
package richtext

import (
	"html"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mq/api/internal/port/outbound"
)

// Sanitizer allows paragraphs, line breaks, two heading levels, bold, italic, underline,
// strikethrough, lists, blockquotes and http(s)/mailto links, and nothing else: no images,
// tables, styles, classes, ids or scripts. It must match what the web editor can produce.
type Sanitizer struct {
	policy *bluemonday.Policy
	strict *bluemonday.Policy
}

func NewSanitizer() outbound.RichTextSanitizer {
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "em", "ul", "ol", "li", "blockquote")
	// Headings are limited to h2/h3: the page title is the h1, and deeper levels add nothing here.
	p.AllowElements("h2", "h3", "u", "s")
	// The editor may emit <b>/<i>; fold them into the canonical tags.
	p.AllowElements("b", "i")
	p.AllowStandardURLs()
	p.AllowAttrs("href").OnElements("a")
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)
	p.AddTargetBlankToFullyQualifiedLinks(true)
	p.AllowURLSchemes("http", "https", "mailto")
	p.RequireParseableURLs(true)
	return &Sanitizer{policy: p, strict: bluemonday.StrictPolicy()}
}

func (s *Sanitizer) Sanitize(in string) (string, string) {
	clean := strings.TrimSpace(s.policy.Sanitize(in))
	// StrictPolicy strips every tag but leaves entities escaped; decode them so the length
	// check counts what a reader sees, not "&amp;".
	text := strings.TrimSpace(html.UnescapeString(s.strict.Sanitize(clean)))
	if text == "" {
		return "", ""
	}
	return clean, text
}
