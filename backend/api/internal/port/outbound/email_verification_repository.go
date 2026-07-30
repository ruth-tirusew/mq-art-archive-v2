package outbound

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type EmailVerificationToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type EmailVerificationRepository interface {
	Create(ctx context.Context, token EmailVerificationToken) error
	Consume(ctx context.Context, tokenHash string, now time.Time) (uuid.UUID, error)
}
