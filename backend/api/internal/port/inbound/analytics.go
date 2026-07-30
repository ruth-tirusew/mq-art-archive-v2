package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/analytics"
)

type AnalyticsService interface {
	RecordView(ctx context.Context, visitorID, entityType string, entityID uuid.UUID) (bool, error)
	Query(ctx context.Context, entityType string, entityID *uuid.UUID, from, to time.Time) ([]analytics.View, error)
	MeOverview(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]analytics.View, error)
}
