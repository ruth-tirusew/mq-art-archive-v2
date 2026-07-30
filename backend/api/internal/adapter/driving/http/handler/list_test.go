package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

func TestProfileHandler_List_success(t *testing.T) {
	h := NewProfileHandler(&mockProfile{
		listApproved: func(ctx context.Context, filter profile.ListFilter) ([]profile.ArtistProfile, error) {
			assist.Equal(t, "painter", filter.Query)
			assist.Equal(t, 10, filter.Limit)
			return []profile.ArtistProfile{{ID: uuid.New(), Slug: "artist", DisplayName: "Artist"}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/artists?q=painter&limit=10", nil, nil, nil, h.List)
	assist.Equal(t, http.StatusOK, w.Code)
	var resp []dto.ArtistProfileResponse
	assist.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assist.Len(t, 1, len(resp))
}

func TestProfileHandler_GetByHandle_success(t *testing.T) {
	h := NewProfileHandler(&mockProfile{
		getArtistByHandle: func(ctx context.Context, handle string) (*profile.ArtistProfile, error) {
			assist.Equal(t, "studio", handle)
			return &profile.ArtistProfile{Handle: handle, DisplayName: "Studio"}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/profiles/@studio", nil,
		gin.Params{{Key: "handle", Value: "studio"}}, nil, h.GetByHandle)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestArtHandler_List_success(t *testing.T) {
	h := NewArtHandler(&mockArt{
		listPublished: func(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
			assist.Equal(t, "Addis", filter.City)
			return []art.ArtPostWithArtist{{
				ArtPost:    art.ArtPost{ID: uuid.New(), Title: "Work", CreatedAt: time.Now().UTC()},
				ArtistSlug: "artist",
				ArtistName: "Artist",
			}}, nil
		},
	}, &mockProfile{})

	w := serve(t, http.MethodGet, "/api/v1/posts?city=Addis", nil, nil, nil, h.List)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestEventHandler_List_success(t *testing.T) {
	h := NewEventHandler(&mockEvent{
		list: func(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
			assist.Equal(t, 2, len(filter.Statuses))
			assist.Equal(t, true, filter.UpcomingOnly)
			return []events.Event{{ID: uuid.New(), Title: "Opening", StartsAt: time.Now().UTC()}}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/events?type=exhibition", nil, nil, nil, h.List)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestEventHandler_List_upcomingFalse(t *testing.T) {
	h := NewEventHandler(&mockEvent{
		list: func(ctx context.Context, filter events.ListFilter) ([]events.Event, error) {
			assist.Equal(t, false, filter.UpcomingOnly)
			return []events.Event{}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/events?upcoming=false", nil, nil, nil, h.List)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestEventHandler_GetBySlug_success(t *testing.T) {
	h := NewEventHandler(&mockEvent{
		getBySlug: func(ctx context.Context, slug string) (*events.Event, error) {
			assist.Equal(t, "opening", slug)
			return &events.Event{Slug: slug, Title: "Opening", StartsAt: time.Now().UTC()}, nil
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/events/opening", nil,
		gin.Params{{Key: "slug", Value: "opening"}}, nil, h.GetBySlug)
	assist.Equal(t, http.StatusOK, w.Code)
}

func TestEventHandler_GetBySlug_notFound(t *testing.T) {
	h := NewEventHandler(&mockEvent{
		getBySlug: func(ctx context.Context, slug string) (*events.Event, error) {
			return nil, apperrors.ErrNotFound
		},
	})

	w := serve(t, http.MethodGet, "/api/v1/events/missing", nil,
		gin.Params{{Key: "slug", Value: "missing"}}, nil, h.GetBySlug)
	assist.Equal(t, http.StatusNotFound, w.Code)
}

func TestQueryHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?limit=25&offset=5&featured=true&year=2024&upcoming=false", nil)

	assist.Equal(t, 25, queryLimit(c, 50))
	assist.Equal(t, 5, queryOffset(c))
	assist.NotNil(t, queryBool(c, "featured"))
	assist.Equal(t, true, *queryBool(c, "featured"))
	assist.NotNil(t, queryInt(c, "year"))
	assist.Equal(t, 2024, *queryInt(c, "year"))

	filter := eventListFilter(c)
	assist.Equal(t, false, filter.UpcomingOnly)
}
