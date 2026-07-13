package art

import "github.com/google/uuid"

type ListFilter struct {
	ArtistID            *uuid.UUID
	City                string
	Medium              string
	Year                *int
	Style               string
	FeaturedAcquisition *bool
	Query               string
	Limit               int
	Offset              int
}

func PublicListFilter() ListFilter {
	return ListFilter{Limit: 50}
}
