package outbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
)

// QuestionnaireArtistLookup is the one profile capability the public answers view needs. The
// profile repository satisfies it.
type QuestionnaireArtistLookup interface {
	GetArtistBySlug(ctx context.Context, slug string) (*profile.ArtistProfile, error)
}

// QuestionnaireApplicationLookup is the one onboarding capability the reviewer's answers view
// needs. The onboarding repository satisfies it.
type QuestionnaireApplicationLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
}
