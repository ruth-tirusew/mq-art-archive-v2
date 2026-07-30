package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/testutil/assist"
)

type mockSearch struct {
	search func(ctx context.Context, query string, limit int) (*inbound.SearchResults, error)
}

func (m *mockSearch) Search(ctx context.Context, query string, limit int) (*inbound.SearchResults, error) {
	return m.search(ctx, query, limit)
}

func TestSearchHandler_missingQuery(t *testing.T) {
	h := NewSearchHandler(&mockSearch{})
	w := serve(t, http.MethodGet, "/api/v1/search", nil, nil, nil, h.Search)
	assist.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSearchHandler_success(t *testing.T) {
	h := NewSearchHandler(&mockSearch{
		search: func(ctx context.Context, query string, limit int) (*inbound.SearchResults, error) {
			assist.Equal(t, "ceramics", query)
			return &inbound.SearchResults{
				Articles: []content.Article{{Slug: "a", Title: "A", Status: content.ArticleStatusPublished}},
				Events:   []events.Event{{Slug: "e", Title: "E", Status: events.EventStatusApproved}},
			}, nil
		},
	})
	w := serve(t, http.MethodGet, "/api/v1/search?q=ceramics", nil, nil, nil, h.Search)
	assist.Equal(t, http.StatusOK, w.Code)
}
