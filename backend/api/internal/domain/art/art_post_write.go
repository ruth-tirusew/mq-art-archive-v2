package art

// ArtPostWrite is the mutable content of a post (create/update).
type ArtPostWrite struct {
	Title       string
	Description string
	Medium      string
	Year        *int
	Dimensions  string
	City        string
	Style       string
	Residency   string
	Exhibition  string
	Palette     []string
	MediaURLs   []string
}
