package inbound

import (
	"context"

	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/profile"
)

type SearchResults struct {
	Articles []content.Article
	Events   []events.Event
	Artists  []profile.ArtistProfile
	Posts    []art.ArtPostWithArtist
}

type SearchService interface {
	Search(ctx context.Context, query string, limit int) (*SearchResults, error)
}
