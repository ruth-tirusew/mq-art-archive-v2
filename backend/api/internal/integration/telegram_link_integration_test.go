//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func TestTelegramLinkRepository_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	repo := postgres.NewTelegramLinkRepository(app.Pool)

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleArtist)

	now := time.Now().UTC()
	token := digest.LinkToken{
		Token:     "test-token-" + uuid.New().String(),
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute),
	}
	assist.NoError(t, repo.Create(ctx, token))

	consumed, err := repo.Consume(ctx, token.Token)
	assist.NoError(t, err)
	assist.Equal(t, userID, consumed.UserID)

	// A second consume of the same token must fail — it's one-time use.
	_, err = repo.Consume(ctx, token.Token)
	assist.Error(t, err)
}

func TestTelegramLinkRepository_integration_expiredTokenRejected(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	repo := postgres.NewTelegramLinkRepository(app.Pool)

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleArtist)

	now := time.Now().UTC()
	token := digest.LinkToken{
		Token:     "expired-token-" + uuid.New().String(),
		UserID:    userID,
		CreatedAt: now.Add(-time.Hour),
		ExpiresAt: now.Add(-time.Minute),
	}
	assist.NoError(t, repo.Create(ctx, token))

	_, err := repo.Consume(ctx, token.Token)
	assist.Error(t, err)
}

func TestNotificationPreferencesRepository_telegramChatID_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	repo := postgres.NewNotificationPreferencesRepository(app.Pool)

	userID := uuid.New()
	integration.InsertUser(t, app.Pool, userID, identity.RoleArtist)

	chatID := "123456789"
	assist.NoError(t, repo.SetTelegramChatID(ctx, userID, &chatID))

	byChat, err := repo.GetByTelegramChatID(ctx, chatID)
	assist.NoError(t, err)
	assist.Equal(t, userID, byChat.UserID)

	byUser, err := repo.GetByUserID(ctx, userID)
	assist.NoError(t, err)
	assist.NotNil(t, byUser.TelegramChatID)
	assist.Equal(t, chatID, *byUser.TelegramChatID)

	// Unlinking clears it from both lookup directions.
	assist.NoError(t, repo.SetTelegramChatID(ctx, userID, nil))
	_, err = repo.GetByTelegramChatID(ctx, chatID)
	assist.Error(t, err)
}
