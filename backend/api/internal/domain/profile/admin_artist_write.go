package profile

// AdminArtistWrite is the mutable content an admin may set on create/update.
type AdminArtistWrite struct {
	Email             string
	DisplayName       string
	Slug              string
	Handle            string
	Bio               string
	Born              string
	Discipline        string
	Tagline           string
	YearsActive       string
	PortraitURL       string
	Influences        []string
	InResidence       bool
	ResidencyPlace    string
	OpenForCommission bool
	Contact           ContactInfo
	Social            SocialLinks
	Status            ProfileStatus
}
