package dto

import (
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/art"
)

type MediaAssetResponse struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	MimeType  string    `json:"mime_type,omitempty"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	SortOrder int       `json:"sort_order"`
}

type ArtPostResponse struct {
	ID                  uuid.UUID            `json:"id"`
	ArtistID            uuid.UUID            `json:"artist_id"`
	ArtistSlug          string               `json:"artist_slug,omitempty"`
	ArtistName          string               `json:"artist_name,omitempty"`
	Title               string               `json:"title"`
	Description         string               `json:"description,omitempty"`
	Medium              string               `json:"medium,omitempty"`
	Year                *int                 `json:"year,omitempty"`
	Dimensions          string               `json:"dimensions,omitempty"`
	City                string               `json:"city,omitempty"`
	Style               string               `json:"style,omitempty"`
	Residency           string               `json:"residency,omitempty"`
	Exhibition          string               `json:"exhibition,omitempty"`
	FeaturedAcquisition bool                 `json:"featured_acquisition"`
	Palette             []string             `json:"palette,omitempty"`
	Media               []MediaAssetResponse `json:"media"`
	Status              string               `json:"status"`
	PublishedAt         *time.Time           `json:"published_at,omitempty"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

func ToMediaAssetResponses(media []domain.MediaAsset) []MediaAssetResponse {
	out := make([]MediaAssetResponse, len(media))
	for i, m := range media {
		out[i] = MediaAssetResponse{
			ID:        m.ID,
			URL:       m.URL,
			MimeType:  m.MimeType,
			Width:     m.Width,
			Height:    m.Height,
			SortOrder: m.SortOrder,
		}
	}
	return out
}

func ToArtPostResponse(post domain.ArtPost) ArtPostResponse {
	return ArtPostResponse{
		ID:                  post.ID,
		ArtistID:            post.ArtistID,
		Title:               post.Title,
		Description:         post.Description,
		Medium:              post.Medium,
		Year:                post.Year,
		Dimensions:          post.Dimensions,
		City:                post.City,
		Style:               post.Style,
		Residency:           post.Residency,
		Exhibition:          post.Exhibition,
		FeaturedAcquisition: post.FeaturedAcquisition,
		Palette:             post.Palette,
		Media:               ToMediaAssetResponses(post.Media),
		Status:              string(post.Status),
		PublishedAt:         post.PublishedAt,
		CreatedAt:           post.CreatedAt,
		UpdatedAt:           post.UpdatedAt,
	}
}

func ToArtPostWithArtistResponse(post domain.ArtPostWithArtist) ArtPostResponse {
	resp := ToArtPostResponse(post.ArtPost)
	resp.ArtistSlug = post.ArtistSlug
	resp.ArtistName = post.ArtistName
	return resp
}

func ToArtPostResponses(posts []domain.ArtPost) []ArtPostResponse {
	out := make([]ArtPostResponse, len(posts))
	for i, p := range posts {
		out[i] = ToArtPostResponse(p)
	}
	return out
}

func ToArtPostWithArtistResponses(posts []domain.ArtPostWithArtist) []ArtPostResponse {
	out := make([]ArtPostResponse, len(posts))
	for i, p := range posts {
		out[i] = ToArtPostWithArtistResponse(p)
	}
	return out
}

type CreateArtPostRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Medium      string   `json:"medium"`
	Year        *int     `json:"year"`
	Dimensions  string   `json:"dimensions"`
	City        string   `json:"city"`
	Style       string   `json:"style"`
	Residency   string   `json:"residency"`
	Exhibition  string   `json:"exhibition"`
	Palette     []string `json:"palette"`
	MediaURLs   []string `json:"media_urls"`
}

func (r CreateArtPostRequest) ToWrite() domain.ArtPostWrite {
	return domain.ArtPostWrite{
		Title:       r.Title,
		Description: r.Description,
		Medium:      r.Medium,
		Year:        r.Year,
		Dimensions:  r.Dimensions,
		City:        r.City,
		Style:       r.Style,
		Residency:   r.Residency,
		Exhibition:  r.Exhibition,
		Palette:     r.Palette,
		MediaURLs:   r.MediaURLs,
	}
}

type UpdateArtPostRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Medium      string   `json:"medium"`
	Year        *int     `json:"year"`
	Dimensions  string   `json:"dimensions"`
	City        string   `json:"city"`
	Style       string   `json:"style"`
	Residency   string   `json:"residency"`
	Exhibition  string   `json:"exhibition"`
	Palette     []string `json:"palette"`
	MediaURLs   []string `json:"media_urls"`
}

func (r UpdateArtPostRequest) ToWrite() domain.ArtPostWrite {
	return domain.ArtPostWrite{
		Title:       r.Title,
		Description: r.Description,
		Medium:      r.Medium,
		Year:        r.Year,
		Dimensions:  r.Dimensions,
		City:        r.City,
		Style:       r.Style,
		Residency:   r.Residency,
		Exhibition:  r.Exhibition,
		Palette:     r.Palette,
		MediaURLs:   r.MediaURLs,
	}
}

type AdminUpdateArtPostRequest struct {
	Status              *string `json:"status"`
	FeaturedAcquisition *bool   `json:"featured_acquisition"`
}

type AdminCreateArtPostRequest struct {
	ArtistID    uuid.UUID `json:"artist_id" binding:"required"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Medium      string    `json:"medium"`
	Year        *int      `json:"year"`
	Dimensions  string    `json:"dimensions"`
	City        string    `json:"city"`
	Style       string    `json:"style"`
	Residency   string    `json:"residency"`
	Exhibition  string    `json:"exhibition"`
	Palette     []string  `json:"palette"`
	MediaURLs   []string  `json:"media_urls"`
	Status      *string   `json:"status"`
}

func (r AdminCreateArtPostRequest) ToWrite() domain.ArtPostWrite {
	return domain.ArtPostWrite{
		Title:       r.Title,
		Description: r.Description,
		Medium:      r.Medium,
		Year:        r.Year,
		Dimensions:  r.Dimensions,
		City:        r.City,
		Style:       r.Style,
		Residency:   r.Residency,
		Exhibition:  r.Exhibition,
		Palette:     r.Palette,
		MediaURLs:   r.MediaURLs,
	}
}
