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
)

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
	eventSource := eventsadapter.NewEventSource(cfg)
	eventsSvc := eventsuc.NewService(eventRepo, eventLocationRepo, eventSource)

	runSync := func() {
		syncCtx, syncCancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer syncCancel()
		n, err := eventsSvc.Sync(syncCtx)
		if err != nil {
			log.Printf("sync error: %v", err)
			return
		}
		log.Printf("sync complete: upserted=%d", n)
	}

	runSync()

	interval := cfg.ScrapeInterval
	if interval <= 0 {
		interval = 6 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			runSync()
		case <-quit:
			log.Printf("scraper shutting down")
			return
		}
	}
}
