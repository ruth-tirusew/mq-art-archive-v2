package analytics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/analytics"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	repo outbound.AnalyticsRepository
	now  func() time.Time
}

func NewService(repo outbound.AnalyticsRepository) inbound.AnalyticsService {
	return &Service{repo: repo, now: time.Now}
}

func (s *Service) RecordView(ctx context.Context, visitorID, entityType string, entityID uuid.UUID) (bool, error) {
	if visitorID == "" || (entityType != "artist" && entityType != "post" && entityType != "article") {
		return false, fmt.Errorf("%w: invalid view", apperrors.ErrValidation)
	}
	now := s.now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	sum := sha256.Sum256([]byte(visitorID + ":" + entityType + ":" + entityID.String() + ":" + day.Format("2006-01-02")))
	return s.repo.RecordUnique(ctx, hex.EncodeToString(sum[:]), entityType, entityID, day, day.Add(48*time.Hour))
}

func (s *Service) Query(ctx context.Context, entityType string, entityID *uuid.UUID, from, to time.Time) ([]analytics.View, error) {
	if to.Before(from) || to.Sub(from) > 366*24*time.Hour {
		return nil, fmt.Errorf("%w: invalid date range", apperrors.ErrValidation)
	}
	return s.repo.Query(ctx, entityType, entityID, from, to)
}

func (s *Service) MeOverview(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]analytics.View, error) {
	if to.Before(from) || to.Sub(from) > 366*24*time.Hour {
		return nil, fmt.Errorf("%w: invalid date range", apperrors.ErrValidation)
	}
	return s.repo.MeOverview(ctx, userID, from, to)
}
