package search

import (
	"context"
	"strings"

	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	articles outbound.ArticleRepository
	events   outbound.EventRepository
	profiles outbound.ProfileRepository
	posts    outbound.ArtPostRepository
}

func NewService(articles outbound.ArticleRepository, events outbound.EventRepository, extras ...any) inbound.SearchService {
	s := &Service{articles: articles, events: events}
	for _, extra := range extras {
		if repo, ok := extra.(outbound.ProfileRepository); ok {
			s.profiles = repo
		}
		if repo, ok := extra.(outbound.ArtPostRepository); ok {
			s.posts = repo
		}
	}
	return s
}

func (s *Service) Search(ctx context.Context, query string, limit int) (*inbound.SearchResults, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return &inbound.SearchResults{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	articles, err := s.articles.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	events, err := s.events.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	results := &inbound.SearchResults{Articles: articles, Events: events}
	if s.profiles != nil {
		results.Artists, err = s.profiles.ListApproved(ctx, profile.ListFilter{Query: query, Limit: limit})
		if err != nil {
			return nil, err
		}
	}
	if s.posts != nil {
		results.Posts, err = s.posts.ListPublished(ctx, art.ListFilter{Query: query, Limit: limit})
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}
