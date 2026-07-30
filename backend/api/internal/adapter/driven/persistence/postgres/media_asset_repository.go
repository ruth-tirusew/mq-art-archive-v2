package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/media"
	"github.com/mq/api/internal/port/outbound"
)

type MediaAssetRepository struct{ pool *Pool }

func NewMediaAssetRepository(pool *Pool) outbound.MediaAssetRepository {
	return &MediaAssetRepository{pool: pool}
}

const mediaAssetColumns = `id, owner_user_id, public_id, secure_url, resource_type, width, height, bytes, COALESCE(folder,''), created_at, updated_at`

func (r *MediaAssetRepository) Create(ctx context.Context, a media.Asset) (*media.Asset, error) {
	row := r.pool.QueryRow(ctx, `INSERT INTO media_assets
		(id,owner_user_id,public_id,secure_url,resource_type,width,height,bytes,folder,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING `+mediaAssetColumns,
		a.ID, a.OwnerUserID, a.PublicID, a.SecureURL, a.ResourceType, a.Width, a.Height, a.Bytes, a.Folder, a.CreatedAt, a.UpdatedAt)
	return scanMediaAsset(row)
}

func (r *MediaAssetRepository) GetByID(ctx context.Context, id uuid.UUID) (*media.Asset, error) {
	return scanMediaAsset(r.pool.QueryRow(ctx, `SELECT `+mediaAssetColumns+` FROM media_assets WHERE id=$1`, id))
}

func (r *MediaAssetRepository) GetByPublicID(ctx context.Context, publicID string) (*media.Asset, error) {
	return scanMediaAsset(r.pool.QueryRow(ctx, `SELECT `+mediaAssetColumns+` FROM media_assets WHERE public_id=$1`, publicID))
}

func (r *MediaAssetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM media_assets WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("delete media asset: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanMediaAsset(row scannable) (*media.Asset, error) {
	var a media.Asset
	err := row.Scan(&a.ID, &a.OwnerUserID, &a.PublicID, &a.SecureURL, &a.ResourceType, &a.Width, &a.Height, &a.Bytes, &a.Folder, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan media asset: %w", err)
	}
	return &a, nil
}
