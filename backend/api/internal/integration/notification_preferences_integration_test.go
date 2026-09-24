//go:build integration

package integration_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

// TestListEventSummaryRecipients_integration guards against the recipients query
// pointing at a table name ("notification_preferences") that doesn't exist — the
// real table is "user_notification_preferences". A fake-repository unit test can't
// catch this class of bug; it needs to run against real Postgres.
func TestListEventSummaryRecipients_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	userRepo := postgres.NewUserRepository(app.Pool)
	// NewNotificationPreferencesRepository's declared return type only carries
	// Get/Upsert; ListEventSummaryRecipients is exposed through the separate
	// EventNotificationRepository interface the same concrete type also satisfies
	// (mirrors how cmd/api's events usecase picks it up via a type switch).
	notifPrefsRepo := postgres.NewNotificationPreferencesRepository(app.Pool)
	notifRepo, ok := notifPrefsRepo.(outbound.EventNotificationRepository)
	if !ok {
		t.Fatal("NotificationPreferencesRepository does not implement EventNotificationRepository")
	}

	optedIn := identity.User{
		ID:        uuid.New(),
		Email:     "recipient-" + uuid.New().String() + "@example.com",
		Role:      identity.RoleArtist,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	assist.NoError(t, userRepo.Create(ctx, optedIn))
	assist.NoError(t, notifPrefsRepo.Upsert(ctx, identity.NotificationPreferences{
		UserID:                  optedIn.ID,
		EmailOnEventSyncSummary: true,
	}))

	optedOut := identity.User{
		ID:        uuid.New(),
		Email:     "not-a-recipient-" + uuid.New().String() + "@example.com",
		Role:      identity.RoleArtist,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	assist.NoError(t, userRepo.Create(ctx, optedOut))
	assist.NoError(t, notifPrefsRepo.Upsert(ctx, identity.NotificationPreferences{
		UserID:                  optedOut.ID,
		EmailOnEventSyncSummary: false,
		NewsletterEnabled:       false,
	}))

	emails, err := notifRepo.ListEventSummaryRecipients(ctx)
	assist.NoError(t, err)
	if !slices.Contains(emails, optedIn.Email) {
		t.Fatalf("expected %q in recipients, got %v", optedIn.Email, emails)
	}
	if slices.Contains(emails, optedOut.Email) {
		t.Fatalf("did not expect %q in recipients, got %v", optedOut.Email, emails)
	}
}
