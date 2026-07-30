package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/analytics"
)

type AnalyticsRepository interface {
	RecordUnique(ctx context.Context, hash, entityType string, entityID uuid.UUID, day, expiresAt time.Time) (bool, error)
	Query(ctx context.Context, entityType string, entityID *uuid.UUID, from, to time.Time) ([]analytics.View, error)
	MeOverview(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]analytics.View, error)
}
