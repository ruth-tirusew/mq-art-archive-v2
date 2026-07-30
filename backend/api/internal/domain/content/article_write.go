package content

// ArticleWrite is the mutable content of a wiki article (create/update).
type ArticleWrite struct {
	Title      string
	Body       string
	Slug       string // optional; when set on draft update, replaces slug
	Category   string
	Excerpt    string
	Difficulty string
	Verified   bool
	Status     *ArticleStatus
}
