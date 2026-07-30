//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestOAuthAccountRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	userRepo := postgres.NewUserRepository(pool)
	oauthRepo := postgres.NewOAuthAccountRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	err := userRepo.Create(ctx, identity.User{
		ID:        userID,
		Email:     "oauth-user@example.com",
		Role:      identity.RolePublic,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	assist.NoError(t, err)

	account := identity.OAuthAccount{
		ID:             uuid.New(),
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: "google-sub-123",
		Email:          "oauth-user@example.com",
		CreatedAt:      time.Now().UTC(),
	}
	assist.NoError(t, oauthRepo.Create(ctx, account))

	got, err := oauthRepo.GetByProviderSubject(ctx, "google", "google-sub-123")
	assist.NoError(t, err)
	assist.Equal(t, userID, got.UserID)
	assist.Equal(t, "google-sub-123", got.ProviderUserID)

	_, err = oauthRepo.GetByProviderSubject(ctx, "google", "missing")
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
