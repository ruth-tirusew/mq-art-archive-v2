//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/institution"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestInstitutionRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewInstitutionRepository(pool)
	ctx := context.Background()

	userID := uuid.New()
	now := time.Now().UTC()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, userID.String()+"@example.com", string(identity.RoleInstitution), now, now)
	assist.NoError(t, err)

	created, err := repo.Save(ctx, institution.Institution{
		ID:          uuid.New(),
		UserID:      userID,
		Slug:        "national-gallery",
		Name:        "National Gallery",
		Description: "Art institution",
		Contact: institution.ContactInfo{
			Location: "Addis Ababa",
		},
		Status:    institution.StatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	})
	assist.NoError(t, err)
	assist.Equal(t, "national-gallery", created.Slug)

	bySlug, err := repo.GetBySlug(ctx, "national-gallery")
	assist.NoError(t, err)
	assist.Equal(t, "National Gallery", bySlug.Name)

	byID, err := repo.GetByID(ctx, created.ID)
	assist.NoError(t, err)
	assist.Equal(t, created.ID, byID.ID)

	created.Name = "National Gallery Updated"
	updated, err := repo.Save(ctx, *created)
	assist.NoError(t, err)
	assist.Equal(t, "National Gallery Updated", updated.Name)

	_, err = repo.GetBySlug(ctx, "missing")
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}
