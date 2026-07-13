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
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Medium      string `json:"medium"`
}
