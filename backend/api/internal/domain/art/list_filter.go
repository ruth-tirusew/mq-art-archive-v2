package art

import (
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	ArtistID            *uuid.UUID
	City                string
	Medium              string
	Year                *int
	Style               string
	FeaturedAcquisition *bool
	Query               string
	// PublishedSince, when set, restricts results to posts whose PublishedAt is at or
	// after this time. Nil means no lower bound.
	PublishedSince *time.Time
	Limit          int
	Offset         int
}

func PublicListFilter() ListFilter {
	return ListFilter{Limit: 50}
}
