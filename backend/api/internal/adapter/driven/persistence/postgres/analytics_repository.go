package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/analytics"
	"github.com/mq/api/internal/port/outbound"
)

type AnalyticsRepository struct{ pool *Pool }

func NewAnalyticsRepository(pool *Pool) outbound.AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) RecordUnique(ctx context.Context, hash, entityType string, entityID uuid.UUID, day, expiresAt time.Time) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	_, _ = tx.Exec(ctx, `DELETE FROM page_view_dedupe WHERE expires_at < NOW()`)
	tag, err := tx.Exec(ctx, `INSERT INTO page_view_dedupe(hash,expires_at) VALUES($1,$2) ON CONFLICT DO NOTHING`, hash, expiresAt)
	if err != nil {
		return false, fmt.Errorf("dedupe page view: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `INSERT INTO page_view_daily(entity_type,entity_id,day,count) VALUES($1,$2,$3,1)
		ON CONFLICT(entity_type,entity_id,day) DO UPDATE SET count=page_view_daily.count+1`, entityType, entityID, day)
	if err != nil {
		return false, fmt.Errorf("record page view: %w", err)
	}
	return true, tx.Commit(ctx)
}

func (r *AnalyticsRepository) Query(ctx context.Context, entityType string, entityID *uuid.UUID, from, to time.Time) ([]analytics.View, error) {
	query := `SELECT entity_type,entity_id,day,count FROM page_view_daily
		WHERE entity_type=$1 AND entity_id=$2 AND day BETWEEN $3 AND $4 ORDER BY day`
	args := []any{entityType, uuid.Nil, from, to}
	if entityID == nil {
		query = `SELECT COALESCE(NULLIF($1, ''), 'all'), '00000000-0000-0000-0000-000000000000'::uuid, day, SUM(count)
			FROM page_view_daily
			WHERE ($1 = '' OR entity_type=$1) AND day BETWEEN $2 AND $3
			GROUP BY day ORDER BY day`
		args = []any{entityType, from, to}
	} else {
		args[1] = *entityID
	}
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []analytics.View{}
	for rows.Next() {
		var v analytics.View
		if err := rows.Scan(&v.EntityType, &v.EntityID, &v.Day, &v.Count); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) MeOverview(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]analytics.View, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT v.entity_type, v.entity_id, v.day, v.count
		FROM page_view_daily v
		WHERE v.day BETWEEN $2 AND $3 AND (
			(v.entity_type = 'artist' AND EXISTS (
				SELECT 1 FROM artist_profiles ap WHERE ap.id = v.entity_id AND ap.user_id = $1
			)) OR
			(v.entity_type = 'post' AND EXISTS (
				SELECT 1 FROM art_posts p
				JOIN artist_profiles ap ON ap.id = p.artist_id
				WHERE p.id = v.entity_id AND ap.user_id = $1
			))
		)
		ORDER BY v.day, v.entity_type, v.entity_id`, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []analytics.View{}
	for rows.Next() {
		var v analytics.View
		if err := rows.Scan(&v.EntityType, &v.EntityID, &v.Day, &v.Count); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}
