//go:build integration

package integration

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mq/api/config"
	store "github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupPostgresPool(t *testing.T) *store.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("mq"),
		tcpostgres.WithUsername("mq"),
		tcpostgres.WithPassword("mq"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	assist.NoError(t, err)

	t.Cleanup(func() {
		assist.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	assist.NoError(t, err)

	RunMigrations(t, connStr)

	pool, err := store.NewPool(ctx, config.Config{DatabaseURL: connStr})
	assist.NoError(t, err)

	t.Cleanup(pool.Close)

	return pool
}

func RunMigrations(t *testing.T, connStr string) {
	t.Helper()

	db, err := sql.Open("pgx", connStr)
	assist.NoError(t, err)
	defer db.Close()

	assist.NoError(t, goose.SetDialect("postgres"))
	assist.NoError(t, goose.Up(db, migrationsDir()))
}

func migrationsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
}
