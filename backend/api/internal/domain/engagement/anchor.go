package engagement

// Anchor pins a comment or highlight to a span of an article's plain-text body, using the
// quote/offset strategy from the wiki-collaboration plan — not stable block IDs. Storing
// the quoted text plus its offsets lets both features ship with no change to the article
// content model, and gives a clean fallback when the underlying text has since changed.
//
// TextOffsetStart/End are RUNE offsets, not byte offsets — they must match the frontend's
// indexing (JS string indices, i.e. UTF-16 code units, which equal rune offsets for every
// character actually used in article bodies here: Latin text, em/en dashes, curly quotes —
// none outside the Basic Multilingual Plane). A body containing a multi-byte UTF-8
// character like "—" would silently misalign anything anchored after it if this were byte
// offsets instead, since Go string indexing/slicing is byte-based by default.
type Anchor struct {
	// QuotedText is the exact substring of the article body this anchor points at, as of
	// when it was created. It's the source of truth for re-verifying the anchor later —
	// the offsets below are a fast-path hint, not something callers should trust blindly.
	QuotedText string
	// TextOffsetStart/End are rune offsets into the article body at anchor time (End
	// exclusive) — see the type doc comment above for why runes, not bytes.
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
	// Matched is true if QuotedText was located in the current body. Start/End (rune
	// offsets, like Anchor's) then give its current position, which can differ from
	// TextOffsetStart/TextOffsetEnd if the article changed before the anchored span.
	// Matched is false when the surrounding text changed enough that the quote can no
	// longer be found — callers should render the comment/highlight unanchored (e.g. in
	// an "on an earlier version" list) rather than guess at a position.
	Matched bool
	Start   int
	End     int
}

// Resolve re-anchors a against an article's current body and version. If the version is
// unchanged, the original offsets are trusted directly without a string search — the
// common case for an unedited article. Otherwise (or if the offsets no longer line up)
// it falls back to a verbatim search for QuotedText. currentBody is converted to runes
// once by the caller's convenience methods below; Resolve itself takes the raw string and
// does the conversion, since a single anchor check is cheap enough not to require callers
// to manage rune slices themselves.
func (a Anchor) Resolve(currentBody string, currentVersion int) Resolution {
	runes := []rune(currentBody)

	if a.ArticleVersionAtAnchor == currentVersion && a.offsetsStillMatch(runes) {
		return Resolution{Anchor: a, Matched: true, Start: a.TextOffsetStart, End: a.TextOffsetEnd}
	}

	if a.QuotedText == "" {
		return Resolution{Anchor: a, Matched: false}
	}
	idx := runeIndex(runes, []rune(a.QuotedText))
	if idx < 0 {
		return Resolution{Anchor: a, Matched: false}
	}
	return Resolution{Anchor: a, Matched: true, Start: idx, End: idx + len([]rune(a.QuotedText))}
}

func (a Anchor) offsetsStillMatch(bodyRunes []rune) bool {
	if a.TextOffsetStart < 0 || a.TextOffsetStart >= a.TextOffsetEnd || a.TextOffsetEnd > len(bodyRunes) {
		return false
	}
	return string(bodyRunes[a.TextOffsetStart:a.TextOffsetEnd]) == a.QuotedText
}

// runeIndex is strings.Index for rune slices — returns the rune offset of the first
// occurrence of needle in haystack, or -1.
func runeIndex(haystack, needle []rune) int {
	if len(needle) == 0 || len(needle) > len(haystack) {
		return -1
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
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

// RuneLen is the rune-counted length callers must use for Coverage's bodyLen and for
// validating a span's bounds — see the Anchor type doc comment for why not len(body).
func RuneLen(body string) int {
	return len([]rune(body))
}

// Slice returns the substring of body spanning the rune offsets [start, end), or "" if
// the span is out of bounds. A small helper so callers validating a client-supplied span
// don't each re-derive the []rune conversion and bounds-checking themselves.
func Slice(body string, start, end int) (string, bool) {
	runes := []rune(body)
	if start < 0 || start >= end || end > len(runes) {
		return "", false
	}
	return string(runes[start:end]), true
}
