// cmd/digest runs the weekly subscriber digest on a fixed weekly schedule (default Monday
// 09:00 UTC, configurable via DIGEST_WEEKDAY/DIGEST_HOUR). It deliberately does not use a
// plain time.Ticker: a ticker fires immediately on process start and drifts from a
// wall-clock target over time, both wrong for a "every Monday at 9am" expectation — see
// docs/operations.md. Instead it computes the next target time and sleeps until then.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mq/api/config"
	maileradapter "github.com/mq/api/internal/adapter/driven/mailer"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	telegramadapter "github.com/mq/api/internal/adapter/driven/telegram"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
	digestuc "github.com/mq/api/internal/usecase/digest"
)

func main() {
	cfg := config.Load()

	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, err := postgres.NewPool(ctx, cfg)
	cancel()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	articleRepo := postgres.NewArticleRepository(pool)
	eventRepo := postgres.NewEventRepository(pool)
	artPostRepo := postgres.NewArtPostRepository(pool)
	digestRecipientRepo := postgres.NewDigestRecipientRepository(pool)
	digestRunRepo := postgres.NewDigestRunRepository(pool)
	notifPrefsRepo := postgres.NewNotificationPreferencesRepository(pool)
	telegramLinkRepo := postgres.NewTelegramLinkRepository(pool)

	var mailer outbound.Mailer = maileradapter.NewLogMailer()
	if cfg.ResendAPIKey != "" {
		mailer = maileradapter.NewResendMailer(cfg.ResendAPIKey, cfg.MailFrom)
	}

	// A nil outbound.TelegramNotifier (not a non-nil interface wrapping a nil *Bot) when
	// the bot isn't configured — digest.Service checks it with a plain != nil.
	var telegramNotifier outbound.TelegramNotifier
	if cfg.TelegramBotToken != "" {
		telegramNotifier = telegramadapter.NewBot(cfg.TelegramBotToken)
	}

	digestSvc := digestuc.NewService(
		articleRepo, eventRepo, artPostRepo, digestRecipientRepo, digestRunRepo,
		notifPrefsRepo, telegramLinkRepo, cfg.JWTSecret,
		mailer, telegramNotifier, cfg.WebAppURL, cfg.PublicAPIURL,
	)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		next := nextWeeklyRun(time.Now().UTC(), cfg.DigestWeekday, cfg.DigestHour)
		log.Printf("digest: next send at %s", next.Format(time.RFC3339))

		select {
		case <-time.After(time.Until(next)):
			runDigest(digestSvc)
		case <-quit:
			log.Printf("digest scheduler shutting down")
			return
		}
	}
}

func runDigest(digestSvc inbound.DigestService) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := digestSvc.SendWeekly(ctx); err != nil {
		log.Printf("digest send error: %v", err)
		return
	}
	log.Printf("digest sent")
}

// nextWeeklyRun returns the next occurrence of weekday at hour:00 UTC strictly after from.
// A run scheduled for right now (from == the target instant) is pushed to next week rather
// than fired immediately, since this function is only ever called to find a future wake-up
// time, never to decide whether "now" itself is a valid run time.
func nextWeeklyRun(from time.Time, weekday time.Weekday, hour int) time.Time {
	from = from.UTC()
	daysUntil := (int(weekday) - int(from.Weekday()) + 7) % 7
	next := time.Date(from.Year(), from.Month(), from.Day(), hour, 0, 0, 0, time.UTC).AddDate(0, 0, daysUntil)
	if !next.After(from) {
		next = next.AddDate(0, 0, 7)
	}
	return next
}
