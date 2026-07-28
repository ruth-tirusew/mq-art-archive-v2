package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/profile"
)

type ContactInfoResponse struct {
	Email    string `json:"email,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Website  string `json:"website,omitempty"`
	Location string `json:"location,omitempty"`
}

type SocialLinksResponse struct {
	Instagram string `json:"instagram,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	Telegram  string `json:"telegram,omitempty"`
}

type ArtistProfileResponse struct {
	ID                uuid.UUID           `json:"id"`
	UserID            uuid.UUID           `json:"user_id"`
	Slug              string              `json:"slug"`
	Handle            string              `json:"handle"`
	DisplayName       string              `json:"display_name"`
	Bio               string              `json:"bio,omitempty"`
	Born              string              `json:"born,omitempty"`
	Discipline        string              `json:"discipline,omitempty"`
	Tagline           string              `json:"tagline,omitempty"`
	YearsActive       string              `json:"years_active,omitempty"`
	PortraitURL       string              `json:"portrait_url,omitempty"`
	Featured          bool                `json:"featured"`
	Influences        []string            `json:"influences"`
	InResidence       bool                `json:"in_residence"`
	ResidencyPlace    string              `json:"residency_place,omitempty"`
	OpenForCommission bool                `json:"open_for_commission"`
	Contact           ContactInfoResponse `json:"contact"`
	Social            SocialLinksResponse `json:"social"`
	Status            string              `json:"status"`
	ApprovedAt        *time.Time          `json:"approved_at,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

func ToArtistProfileResponse(p domain.ArtistProfile) ArtistProfileResponse {
	influences := p.Influences
	if influences == nil {
		influences = []string{}
	}
	return ArtistProfileResponse{
		ID:                p.ID,
		UserID:            p.UserID,
		Slug:              p.Slug,
		Handle:            p.Handle,
		DisplayName:       p.DisplayName,
		Bio:               p.Bio,
		Born:              p.Born,
		Discipline:        p.Discipline,
		Tagline:           p.Tagline,
		YearsActive:       p.YearsActive,
		PortraitURL:       p.PortraitURL,
		Featured:          p.Featured,
		Influences:        influences,
		InResidence:       p.InResidence,
		ResidencyPlace:    p.ResidencyPlace,
		OpenForCommission: p.OpenForCommission,
		Contact: ContactInfoResponse{
			Email:    p.Contact.Email,
			Phone:    p.Contact.Phone,
			Website:  p.Contact.Website,
			Location: p.Contact.Location,
		},
		Social: SocialLinksResponse{
			Instagram: p.Social.Instagram,
			Twitter:   p.Social.Twitter,
			Telegram:  p.Social.Telegram,
		},
		Status:     string(p.Status),
		ApprovedAt: p.ApprovedAt,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func ToArtistProfileResponses(profiles []domain.ArtistProfile) []ArtistProfileResponse {
	out := make([]ArtistProfileResponse, len(profiles))
	for i, p := range profiles {
		out[i] = ToArtistProfileResponse(p)
	}
	return out
}

type ContactInfoRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Website  string `json:"website"`
	Location string `json:"location"`
}

type SocialLinksRequest struct {
	Instagram string `json:"instagram"`
	Twitter   string `json:"twitter"`
	Telegram  string `json:"telegram"`
}

type UpdateArtistProfileRequest struct {
	DisplayName       string             `json:"display_name"`
	Slug              string             `json:"slug"`
	Handle            string             `json:"handle"`
	Bio               string             `json:"bio"`
	Born              string             `json:"born"`
	Discipline        string             `json:"discipline"`
	Tagline           string             `json:"tagline"`
	YearsActive       string             `json:"years_active"`
	PortraitURL       string             `json:"portrait_url"`
	Influences        []string           `json:"influences"`
	InResidence       bool               `json:"in_residence"`
	ResidencyPlace    string             `json:"residency_place"`
	OpenForCommission bool               `json:"open_for_commission"`
	Contact           ContactInfoRequest `json:"contact"`
	Social            SocialLinksRequest `json:"social"`
	Status            string             `json:"status"`
}

func (r UpdateArtistProfileRequest) ToOwnProfileUpdate() domain.OwnProfileUpdate {
	influences := r.Influences
	if influences == nil {
		influences = []string{}
	}
	return domain.OwnProfileUpdate{
		DisplayName:       r.DisplayName,
		Slug:              r.Slug,
		Handle:            r.Handle,
		Bio:               r.Bio,
		Born:              r.Born,
		Discipline:        r.Discipline,
		Tagline:           r.Tagline,
		YearsActive:       r.YearsActive,
		PortraitURL:       r.PortraitURL,
		Influences:        influences,
		InResidence:       r.InResidence,
		ResidencyPlace:    r.ResidencyPlace,
		OpenForCommission: r.OpenForCommission,
		Contact: domain.ContactInfo{
			Email:    r.Contact.Email,
			Phone:    r.Contact.Phone,
			Website:  r.Contact.Website,
			Location: r.Contact.Location,
		},
		Social: domain.SocialLinks{
			Instagram: r.Social.Instagram,
			Twitter:   r.Social.Twitter,
			Telegram:  r.Social.Telegram,
		},
		Status: domain.ProfileStatus(r.Status),
	}
}

type AdminArtistWriteRequest struct {
	Email             string             `json:"email"`
	DisplayName       string             `json:"display_name" binding:"required"`
	Slug              string             `json:"slug"`
	Handle            string             `json:"handle"`
	Bio               string             `json:"bio"`
	Born              string             `json:"born"`
	Discipline        string             `json:"discipline"`
	Tagline           string             `json:"tagline"`
	YearsActive       string             `json:"years_active"`
	PortraitURL       string             `json:"portrait_url"`
	Influences        []string           `json:"influences"`
	InResidence       bool               `json:"in_residence"`
	ResidencyPlace    string             `json:"residency_place"`
	OpenForCommission bool               `json:"open_for_commission"`
	Contact           ContactInfoRequest `json:"contact"`
	Social            SocialLinksRequest `json:"social"`
	Status            string             `json:"status"`
}

func (r AdminArtistWriteRequest) ToWrite() domain.AdminArtistWrite {
	influences := r.Influences
	if influences == nil {
		influences = []string{}
	}
	return domain.AdminArtistWrite{
		Email:             r.Email,
		DisplayName:       r.DisplayName,
		Slug:              r.Slug,
		Handle:            r.Handle,
		Bio:               r.Bio,
		Born:              r.Born,
		Discipline:        r.Discipline,
		Tagline:           r.Tagline,
		YearsActive:       r.YearsActive,
		PortraitURL:       r.PortraitURL,
		Influences:        influences,
		InResidence:       r.InResidence,
		ResidencyPlace:    r.ResidencyPlace,
		OpenForCommission: r.OpenForCommission,
		Contact: domain.ContactInfo{
			Email:    r.Contact.Email,
			Phone:    r.Contact.Phone,
			Website:  r.Contact.Website,
			Location: r.Contact.Location,
		},
		Social: domain.SocialLinks{
			Instagram: r.Social.Instagram,
			Twitter:   r.Social.Twitter,
			Telegram:  r.Social.Telegram,
		},
		Status: domain.ProfileStatus(r.Status),
	}
}
