package outbound

// RichTextSanitizer cleans user-supplied HTML down to a fixed allowlist before it is stored or
// rendered. It is the only thing standing between an artist's answer and every visitor's browser,
// so rich answers must never be saved or served without passing through it.
type RichTextSanitizer interface {
	// Sanitize returns the cleaned HTML and its visible text with all markup removed.
	// Empty markup such as "<p></p>" yields ("", "").
	Sanitize(html string) (clean, text string)
}
