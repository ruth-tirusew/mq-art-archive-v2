package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mq/api/config"
	eventsadapter "github.com/mq/api/internal/adapter/driven/events"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	eventsuc "github.com/mq/api/internal/usecase/events"
	settingsuc "github.com/mq/api/internal/usecase/settings"
)

// checkInterval controls how often the settings row is polled for interval/config
// changes made through the admin UI. It's independent of the scrape interval itself.
const checkInterval = time.Minute

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, err := postgres.NewPool(ctx, cfg)
	cancel()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	eventRepo := postgres.NewEventRepository(pool)
	eventLocationRepo := postgres.NewEventLocationRepository(pool)
	scrapeSettingsRepo := postgres.NewScrapeSettingsRepository(pool)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := settingsuc.EnsureSeed(seedCtx, scrapeSettingsRepo, eventsadapter.SettingsFromConfig(cfg)); err != nil {
		seedCancel()
		log.Fatalf("scrape settings seed: %v", err)
	}
	seedCancel()

	initialCtx, initialCancel := context.WithTimeout(context.Background(), 10*time.Second)
	initialSettings, err := scrapeSettingsRepo.Get(initialCtx)
	initialCancel()
	if err != nil {
		log.Fatalf("scrape settings: %v", err)
	}

	// The event source is rebuilt from the current DB-persisted settings before every
	// run, so an admin toggling scrape/telegram or changing channels/keywords takes
	// effect on the next sync without a restart.
	swapper := eventsadapter.NewSwappableEventSource(
		eventsadapter.NewEventSourceFromSettings(*initialSettings, cfg.TelegramSessionPath),
	)
	reloader := eventsadapter.NewEventSourceReloader(swapper, cfg.TelegramSessionPath)
	eventsSvc := eventsuc.NewService(eventRepo, eventLocationRepo, swapper)

	runSync := func() {
		settingsCtx, settingsCancel := context.WithTimeout(context.Background(), 10*time.Second)
		current, err := scrapeSettingsRepo.Get(settingsCtx)
		settingsCancel()
		if err != nil {
			log.Printf("scrape settings reload error: %v", err)
			return
		}
		if err := reloader.Reload(*current); err != nil {
			log.Printf("event source reload error: %v", err)
			return
		}

		syncCtx, syncCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		n, syncErr := eventsSvc.Sync(syncCtx)
		syncCancel()

		runAt := time.Now().UTC()
		recordCtx, recordCancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := scrapeSettingsRepo.RecordRunResult(recordCtx, runAt, syncErr); err != nil {
			log.Printf("record scrape run result: %v", err)
		}
		recordCancel()

		if syncErr != nil {
			log.Printf("sync error: %v", syncErr)
			return
		}
		log.Printf("sync complete: upserted=%d", n)
	}

	runSync()

	currentInterval := scrapeIntervalOrDefault(initialSettings.ScrapeIntervalSeconds)
	nextRun := time.Now().Add(currentInterval)

	checkTicker := time.NewTicker(checkInterval)
	defer checkTicker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-checkTicker.C:
			settingsCtx, settingsCancel := context.WithTimeout(context.Background(), 10*time.Second)
			current, err := scrapeSettingsRepo.Get(settingsCtx)
			settingsCancel()
			if err != nil {
				log.Printf("scrape settings poll error: %v", err)
				continue
			}
			newInterval := scrapeIntervalOrDefault(current.ScrapeIntervalSeconds)
			if newInterval != currentInterval {
				log.Printf("scrape interval changed: %s -> %s", currentInterval, newInterval)
				currentInterval = newInterval
				nextRun = time.Now().Add(currentInterval)
			}
			if !time.Now().Before(nextRun) {
				runSync()
				nextRun = time.Now().Add(currentInterval)
			}
		case <-quit:
			log.Printf("scraper shutting down")
			return
		}
	}
}

func scrapeIntervalOrDefault(seconds int) time.Duration {
	if seconds <= 0 {
		return 6 * time.Hour
	}
	return time.Duration(seconds) * time.Second
}
