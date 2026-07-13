package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/institution"
)

type InstitutionRepository interface {
	GetBySlug(ctx context.Context, slug string) (*institution.Institution, error)
	GetByID(ctx context.Context, id uuid.UUID) (*institution.Institution, error)
	Save(ctx context.Context, inst institution.Institution) (*institution.Institution, error)
}
