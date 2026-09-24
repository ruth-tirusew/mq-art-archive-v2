package inbound

import (
	"time"

	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
)

// Digest is the content window for one send: events starting soon, plus art posts and
// articles published since the previous run (or the last 7 days, if there's no prior run).
// It lives here rather than in internal/domain/digest because, like SearchResults, it's an
// intentional read-model aggregate across bounded contexts.
type Digest struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Events      []events.Event
	ArtPosts    []art.ArtPostWithArtist
	Articles    []content.Article
}
