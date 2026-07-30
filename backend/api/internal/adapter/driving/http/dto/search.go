package dto

import (
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/profile"
)

type SearchResponse struct {
	Articles []ArticleResponse       `json:"articles"`
	Events   []EventResponse         `json:"events"`
	Artists  []ArtistProfileResponse `json:"artists"`
	Posts    []ArtPostResponse       `json:"posts"`
}

func ToSearchResponse(articles []content.Article, evts []events.Event, artists []profile.ArtistProfile, posts []art.ArtPostWithArtist) SearchResponse {
	return SearchResponse{
		Articles: ToArticleResponses(articles),
		Events:   ToEventResponses(evts),
		Artists:  ToArtistProfileResponses(artists),
		Posts:    ToArtPostWithArtistResponses(posts),
	}
}
