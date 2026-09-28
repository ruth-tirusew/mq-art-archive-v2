package engagement

import "strings"

// Anchor pins a comment or highlight to a span of an article's plain-text body, using the
// quote/offset strategy from the wiki-collaboration plan — not stable block IDs. Storing
// the quoted text plus its offsets lets both features ship with no change to the article
// content model, and gives a clean fallback when the underlying text has since changed.
type Anchor struct {
	// QuotedText is the exact substring of the article body this anchor points at, as of
	// when it was created. It's the source of truth for re-verifying the anchor later —
	// the offsets below are a fast-path hint, not something callers should trust blindly.
	QuotedText string
	// TextOffsetStart/End are character offsets into the article body's plain text at
	// anchor time (End exclusive).
	TextOffsetStart int
	TextOffsetEnd   int
	// ArticleVersionAtAnchor is the article's version when this anchor was created,
	// matching the existing articles.version column. Comparing it to the article's
	// current version is a cheap way to skip re-verification entirely in the common case
	// (nobody has edited the article since).
	ArticleVersionAtAnchor int
}

// Resolution is the outcome of re-checking an Anchor against an article's current body
// and version.
type Resolution struct {
	Anchor
	// Matched is true if QuotedText was located in the current body. Start/End then give
	// its current offsets, which can differ from TextOffsetStart/TextOffsetEnd if the
	// article changed before the anchored span. Matched is false when the surrounding
	// text changed enough that the quote can no longer be found — callers should render
	// the comment/highlight unanchored (e.g. in an "on an earlier version" list) rather
	// than guess at a position.
	Matched bool
	Start   int
	End     int
}

// Resolve re-anchors a against an article's current body and version. If the version is
// unchanged, the original offsets are trusted directly without a string search — the
// common case for an unedited article. Otherwise (or if the offsets no longer line up)
// it falls back to a verbatim search for QuotedText.
func (a Anchor) Resolve(currentBody string, currentVersion int) Resolution {
	if a.ArticleVersionAtAnchor == currentVersion && a.offsetsStillMatch(currentBody) {
		return Resolution{Anchor: a, Matched: true, Start: a.TextOffsetStart, End: a.TextOffsetEnd}
	}

	if a.QuotedText == "" {
		return Resolution{Anchor: a, Matched: false}
	}
	idx := strings.Index(currentBody, a.QuotedText)
	if idx < 0 {
		return Resolution{Anchor: a, Matched: false}
	}
	return Resolution{Anchor: a, Matched: true, Start: idx, End: idx + len(a.QuotedText)}
}

func (a Anchor) offsetsStillMatch(body string) bool {
	if a.TextOffsetStart < 0 || a.TextOffsetStart >= a.TextOffsetEnd || a.TextOffsetEnd > len(body) {
		return false
	}
	return body[a.TextOffsetStart:a.TextOffsetEnd] == a.QuotedText
}

// ResolveAll resolves several anchors against the same article body/version in one pass —
// what a comment or highlight list endpoint needs before returning to the client.
func ResolveAll(anchors []Anchor, currentBody string, currentVersion int) []Resolution {
	out := make([]Resolution, len(anchors))
	for i, a := range anchors {
		out[i] = a.Resolve(currentBody, currentVersion)
	}
	return out
}
