package profile

import (
	"time"

	"github.com/google/uuid"
)

type ProfileStatus string

const (
	ProfileStatusDraft    ProfileStatus = "draft"
	ProfileStatusPending  ProfileStatus = "pending"
	ProfileStatusApproved ProfileStatus = "approved"
)

type ContactInfo struct {
	Email    string
	Phone    string
	Website  string
	Location string
}

type SocialLinks struct {
	Instagram string
	Twitter   string
	Telegram  string
}

type ArtistProfile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Slug        string
	Handle      string
	DisplayName string
	Bio         string
	Born        string
	Discipline  string
	Tagline     string
	YearsActive string
	PortraitURL string
	Featured    bool
	Contact     ContactInfo
	Social      SocialLinks
	Status      ProfileStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
