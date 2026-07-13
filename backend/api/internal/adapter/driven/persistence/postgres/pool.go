package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mq/api/config"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(ctx context.Context, cfg config.Config) (*Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Pool{Pool: pool}, nil
}

func (p *Pool) Close() {
	p.Pool.Close()
}
