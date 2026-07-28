package art

import (
	"time"

	"github.com/google/uuid"
)

type ArtStatus string

const (
	ArtStatusDraft     ArtStatus = "draft"
	ArtStatusPublished ArtStatus = "published"
	ArtStatusArchived  ArtStatus = "archived"
)

type MediaAsset struct {
	ID        uuid.UUID
	URL       string
	MimeType  string
	Width     int
	Height    int
	SortOrder int
}

type ArtPost struct {
	ID                  uuid.UUID
	ArtistID            uuid.UUID
	Title               string
	Description         string
	Medium              string
	Year                *int
	Dimensions          string
	City                string
	Style               string
	Residency           string
	Exhibition          string
	FeaturedAcquisition bool
	Palette             []string
	Media               []MediaAsset
	Status              ArtStatus
	PublishedAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ArtPostWithArtist is used for archive listings.
type ArtPostWithArtist struct {
	ArtPost
	ArtistSlug string
	ArtistName string
}
