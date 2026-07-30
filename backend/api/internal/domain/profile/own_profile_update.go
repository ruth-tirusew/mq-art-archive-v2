package profile

// OwnProfileUpdate contains fields an artist may change on their own profile.
type OwnProfileUpdate struct {
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
