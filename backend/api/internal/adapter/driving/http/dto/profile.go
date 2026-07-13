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
	ID          uuid.UUID            `json:"id"`
	UserID      uuid.UUID            `json:"user_id"`
	Slug        string               `json:"slug"`
	Handle      string               `json:"handle"`
	DisplayName string               `json:"display_name"`
	Bio         string               `json:"bio,omitempty"`
	Born        string               `json:"born,omitempty"`
	Discipline  string               `json:"discipline,omitempty"`
	Tagline     string               `json:"tagline,omitempty"`
	YearsActive string               `json:"years_active,omitempty"`
	PortraitURL string               `json:"portrait_url,omitempty"`
	Featured    bool                 `json:"featured"`
	Contact     ContactInfoResponse  `json:"contact"`
	Social      SocialLinksResponse  `json:"social"`
	Status      string               `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func ToArtistProfileResponse(p domain.ArtistProfile) ArtistProfileResponse {
	return ArtistProfileResponse{
		ID:          p.ID,
		UserID:      p.UserID,
		Slug:        p.Slug,
		Handle:      p.Handle,
		DisplayName: p.DisplayName,
		Bio:         p.Bio,
		Born:        p.Born,
		Discipline:  p.Discipline,
		Tagline:     p.Tagline,
		YearsActive: p.YearsActive,
		PortraitURL: p.PortraitURL,
		Featured:    p.Featured,
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
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ToArtistProfileResponses(profiles []domain.ArtistProfile) []ArtistProfileResponse {
	out := make([]ArtistProfileResponse, len(profiles))
	for i, p := range profiles {
		out[i] = ToArtistProfileResponse(p)
	}
	return out
}
