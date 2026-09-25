//go:build integration

package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
	digestuc "github.com/mq/api/internal/usecase/digest"
)

// spyMailer records every SendHTML call. Send (plain text) isn't used by the digest.
type spyMailer struct {
	mu    sync.Mutex
	calls []string // recipient emails
}

func (m *spyMailer) Send(ctx context.Context, to, subject, body string) error { return nil }

func (m *spyMailer) SendHTML(ctx context.Context, to, subject, htmlBody, textBody string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, to)
	return nil
}

func (m *spyMailer) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// spyTelegram records every Send call.
type spyTelegram struct {
	mu    sync.Mutex
	calls []string // chat IDs
}

func (t *spyTelegram) Send(ctx context.Context, chatID, message string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, chatID)
	return nil
}

func (t *spyTelegram) callCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.calls)
}

func newDigestServiceForTest(app *integration.App, mailer *spyMailer, telegram *spyTelegram) *digestServiceUnderTest {
	articleRepo := postgres.NewArticleRepository(app.Pool)
	eventRepo := postgres.NewEventRepository(app.Pool)
	artPostRepo := postgres.NewArtPostRepository(app.Pool)
	digestRecipientRepo := postgres.NewDigestRecipientRepository(app.Pool)
	digestRunRepo := postgres.NewDigestRunRepository(app.Pool)
	notifPrefsRepo := postgres.NewNotificationPreferencesRepository(app.Pool)
	telegramLinkRepo := postgres.NewTelegramLinkRepository(app.Pool)

	svc := digestuc.NewService(
		articleRepo, eventRepo, artPostRepo, digestRecipientRepo, digestRunRepo,
		notifPrefsRepo, telegramLinkRepo, "test-unsubscribe-secret",
		mailer, telegram, "http://localhost:5173", "http://localhost:8080",
	)
	return &digestServiceUnderTest{svc: svc, runRepo: digestRunRepo, notifRepo: notifPrefsRepo}
}

type digestServiceUnderTest struct {
	svc interface {
		SendWeekly(ctx context.Context) error
	}
	runRepo   outbound.DigestRunRepository
	notifRepo outbound.NotificationPreferencesRepository
}

func TestDigestService_SendWeekly_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	mailer := &spyMailer{}
	telegram := &spyTelegram{}
	under := newDigestServiceForTest(app, mailer, telegram)

	emailUser := uuid.New()
	integration.InsertUser(t, app.Pool, emailUser, identity.RoleArtist)
	assist.NoError(t, under.notifRepo.Upsert(ctx, identity.NotificationPreferences{
		UserID:            emailUser,
		NewsletterEnabled: true,
	}))

	telegramUser := uuid.New()
	integration.InsertUser(t, app.Pool, telegramUser, identity.RoleArtist)
	assist.NoError(t, under.notifRepo.SetTelegramChatID(ctx, telegramUser, ptrStr("555")))

	assist.NoError(t, under.svc.SendWeekly(ctx))

	assist.Equal(t, 1, mailer.callCount())
	assist.Equal(t, 1, telegram.callCount())

	run, err := under.runRepo.LastCompletedRun(ctx)
	assist.NoError(t, err)
	assist.NotNil(t, run.CompletedAt)
}

func TestDigestService_SendWeekly_resumesIncompleteRun_integration(t *testing.T) {
	app := integration.NewApp(t)
	ctx := context.Background()
	mailer := &spyMailer{}
	telegram := &spyTelegram{}
	under := newDigestServiceForTest(app, mailer, telegram)

	alreadySent := uuid.New()
	integration.InsertUser(t, app.Pool, alreadySent, identity.RoleArtist)
	assist.NoError(t, under.notifRepo.Upsert(ctx, identity.NotificationPreferences{
		UserID:            alreadySent,
		NewsletterEnabled: true,
	}))

	stillPending := uuid.New()
	integration.InsertUser(t, app.Pool, stillPending, identity.RoleArtist)
	assist.NoError(t, under.notifRepo.Upsert(ctx, identity.NotificationPreferences{
		UserID:            stillPending,
		NewsletterEnabled: true,
	}))

	// Simulate a process that started a run, successfully emailed one recipient, then
	// crashed before completing — the exact scenario currentOrNewRun exists to handle.
	now := time.Now().UTC()
	run, err := under.runRepo.StartRun(ctx, now.Add(-7*24*time.Hour), now)
	assist.NoError(t, err)
	assist.NoError(t, under.runRepo.RecordDelivery(ctx, run.ID, alreadySent, "email", now, nil))

	assist.NoError(t, under.svc.SendWeekly(ctx))

	// Only the recipient without a prior successful delivery should have been mailed.
	assist.Equal(t, 1, mailer.callCount())

	// The run resumed is the same row, now completed — not a second one.
	rows, err := app.Pool.Query(ctx, `SELECT id, completed_at FROM digest_runs`)
	assist.NoError(t, err)
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id uuid.UUID
		var completedAt *time.Time
		assist.NoError(t, rows.Scan(&id, &completedAt))
		assist.Equal(t, run.ID, id)
		assist.NotNil(t, completedAt)
		count++
	}
	assist.Equal(t, 1, count)
}

func ptrStr(s string) *string { return &s }
