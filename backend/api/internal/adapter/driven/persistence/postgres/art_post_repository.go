package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/port/outbound"
)

type ArtPostRepository struct {
	pool *Pool
}

func NewArtPostRepository(pool *Pool) outbound.ArtPostRepository {
	return &ArtPostRepository{pool: pool}
}

const artPostTableColumns = `id, artist_id, title, description, medium, year, dimensions, city, style, COALESCE(residency,''), COALESCE(exhibition,''), featured_acquisition, palette, status, published_at, created_at, updated_at`

const artPostJoinColumns = `p.id, p.artist_id, p.title, p.description, p.medium, p.year, p.dimensions, p.city, p.style, COALESCE(p.residency,''), COALESCE(p.exhibition,''), p.featured_acquisition, p.palette, p.status, p.published_at, p.created_at, p.updated_at`

func (r *ArtPostRepository) ListByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return r.listByArtist(ctx, artistID, true)
}

func (r *ArtPostRepository) ListOwnedByArtist(ctx context.Context, artistID uuid.UUID) ([]art.ArtPost, error) {
	return r.listByArtist(ctx, artistID, false)
}

func (r *ArtPostRepository) listByArtist(ctx context.Context, artistID uuid.UUID, publishedOnly bool) ([]art.ArtPost, error) {
	query := `
		SELECT ` + artPostTableColumns + `
		FROM art_posts
		WHERE artist_id = $1`
	args := []any{artistID}
	if publishedOnly {
		query += ` AND status = $2`
		args = append(args, string(art.ArtStatusPublished))
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list art posts by artist: %w", err)
	}
	defer rows.Close()

	var posts []art.ArtPost
	for rows.Next() {
		post, err := scanArtPost(rows)
		if err != nil {
			return nil, err
		}
		media, err := r.loadMedia(ctx, post.ID)
		if err != nil {
			return nil, err
		}
		post.Media = media
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list art posts by artist: %w", err)
	}
	if posts == nil {
		posts = []art.ArtPost{}
	}
	return posts, nil
}

func (r *ArtPostRepository) ListPublished(ctx context.Context, filter art.ListFilter) ([]art.ArtPostWithArtist, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT ` + artPostJoinColumns + `, ap.slug, ap.display_name
		FROM art_posts p
		JOIN artist_profiles ap ON ap.id = p.artist_id
		WHERE p.status = $1 AND ap.status = $2`
	args := []any{string(art.ArtStatusPublished), string("approved")}
	argPos := 3

	if filter.ArtistID != nil {
		query += fmt.Sprintf(" AND p.artist_id = $%d", argPos)
		args = append(args, *filter.ArtistID)
		argPos++
	}
	if filter.City != "" {
		query += fmt.Sprintf(" AND p.city ILIKE $%d", argPos)
		args = append(args, filter.City)
		argPos++
	}
	if filter.Medium != "" {
		query += fmt.Sprintf(" AND p.medium ILIKE $%d", argPos)
		args = append(args, "%"+filter.Medium+"%")
		argPos++
	}
	if filter.Year != nil {
		query += fmt.Sprintf(" AND p.year = $%d", argPos)
		args = append(args, *filter.Year)
		argPos++
	}
	if filter.Style != "" {
		query += fmt.Sprintf(" AND p.style ILIKE $%d", argPos)
		args = append(args, "%"+filter.Style+"%")
		argPos++
	}
	if filter.FeaturedAcquisition != nil {
		query += fmt.Sprintf(" AND p.featured_acquisition = $%d", argPos)
		args = append(args, *filter.FeaturedAcquisition)
		argPos++
	}
	if filter.Query != "" {
		query += fmt.Sprintf(" AND (p.title ILIKE $%d OR p.description ILIKE $%d OR p.medium ILIKE $%d OR p.city ILIKE $%d OR p.style ILIKE $%d OR ap.display_name ILIKE $%d)", argPos, argPos, argPos, argPos, argPos, argPos)
		args = append(args, "%"+filter.Query+"%")
		argPos++
	}

	query += " ORDER BY p.published_at DESC NULLS LAST, p.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list published art posts: %w", err)
	}
	defer rows.Close()

	var posts []art.ArtPostWithArtist
	for rows.Next() {
		var item art.ArtPostWithArtist
		var status string
		err := rows.Scan(
			&item.ID, &item.ArtistID, &item.Title, &item.Description, &item.Medium,
			&item.Year, &item.Dimensions, &item.City, &item.Style, &item.Residency, &item.Exhibition, &item.FeaturedAcquisition, &item.Palette,
			&status, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.ArtistSlug, &item.ArtistName,
		)
		if err != nil {
			return nil, err
		}
		item.Status = art.ArtStatus(status)
		media, err := r.loadMedia(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Media = media
		posts = append(posts, item)
	}
	if posts == nil {
		posts = []art.ArtPostWithArtist{}
	}
	return posts, rows.Err()
}

func (r *ArtPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*art.ArtPost, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+artPostTableColumns+`
		FROM art_posts
		WHERE id = $1
	`, id)

	post, err := scanArtPost(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get art post by id: %w", err)
	}

	media, err := r.loadMedia(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Media = media
	return &post, nil
}

func (r *ArtPostRepository) Create(ctx context.Context, post art.ArtPost) (*art.ArtPost, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create art post: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO art_posts (
			id, artist_id, title, description, medium, year, dimensions, city, style, residency, exhibition,
			featured_acquisition, palette, status, published_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`,
		post.ID, post.ArtistID, post.Title, post.Description, post.Medium,
		post.Year, post.Dimensions, post.City, post.Style, emptyToNil(post.Residency), emptyToNil(post.Exhibition),
		post.FeaturedAcquisition, coalescePalette(post.Palette), string(post.Status), post.PublishedAt, post.CreatedAt, post.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create art post: %w", err)
	}

	for _, m := range post.Media {
		mediaID := m.ID
		if mediaID == uuid.Nil {
			mediaID = uuid.New()
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO art_post_media (id, art_post_id, url, mime_type, width, height, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, mediaID, post.ID, m.URL, m.MimeType, m.Width, m.Height, m.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("create art post media: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create art post: commit: %w", err)
	}

	return r.GetByID(ctx, post.ID)
}

func (r *ArtPostRepository) Update(ctx context.Context, post art.ArtPost) (*art.ArtPost, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("update art post: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE art_posts SET
			title = $2, description = $3, medium = $4, year = $5, dimensions = $6,
			city = $7, style = $8, residency = $9, exhibition = $10, featured_acquisition = $11, palette = $12,
			status = $13, published_at = $14, updated_at = $15
		WHERE id = $1
	`,
		post.ID, post.Title, post.Description, post.Medium, post.Year, post.Dimensions,
		post.City, post.Style, emptyToNil(post.Residency), emptyToNil(post.Exhibition), post.FeaturedAcquisition, coalescePalette(post.Palette),
		string(post.Status), post.PublishedAt, post.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update art post: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM art_post_media WHERE art_post_id = $1`, post.ID)
	if err != nil {
		return nil, fmt.Errorf("update art post media clear: %w", err)
	}

	for _, m := range post.Media {
		mediaID := m.ID
		if mediaID == uuid.Nil {
			mediaID = uuid.New()
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO art_post_media (id, art_post_id, url, mime_type, width, height, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, mediaID, post.ID, m.URL, m.MimeType, m.Width, m.Height, m.SortOrder)
		if err != nil {
			return nil, fmt.Errorf("update art post media: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("update art post: commit: %w", err)
	}
	return r.GetByID(ctx, post.ID)
}

func (r *ArtPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("delete art post: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM art_post_media WHERE art_post_id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete art post media: %w", err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM art_posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete art post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("delete art post: commit: %w", err)
	}
	return nil
}

func (r *ArtPostRepository) ListAll(ctx context.Context, status *art.ArtStatus, limit, offset int) ([]art.ArtPostWithArtist, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT ` + artPostJoinColumns + `, ap.slug, ap.display_name
		FROM art_posts p
		JOIN artist_profiles ap ON ap.id = p.artist_id
		WHERE 1=1`
	args := []any{}
	argPos := 1
	if status != nil {
		query += fmt.Sprintf(" AND p.status = $%d", argPos)
		args = append(args, string(*status))
		argPos++
	}
	query += " ORDER BY p.updated_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list all art posts: %w", err)
	}
	defer rows.Close()

	var posts []art.ArtPostWithArtist
	for rows.Next() {
		var item art.ArtPostWithArtist
		var st string
		err := rows.Scan(
			&item.ID, &item.ArtistID, &item.Title, &item.Description, &item.Medium,
			&item.Year, &item.Dimensions, &item.City, &item.Style, &item.Residency, &item.Exhibition, &item.FeaturedAcquisition, &item.Palette,
			&st, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.ArtistSlug, &item.ArtistName,
		)
		if err != nil {
			return nil, err
		}
		item.Status = art.ArtStatus(st)
		media, err := r.loadMedia(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Media = media
		posts = append(posts, item)
	}
	if posts == nil {
		posts = []art.ArtPostWithArtist{}
	}
	return posts, rows.Err()
}

func (r *ArtPostRepository) loadMedia(ctx context.Context, postID uuid.UUID) ([]art.MediaAsset, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, url, mime_type, width, height, sort_order
		FROM art_post_media
		WHERE art_post_id = $1
		ORDER BY sort_order ASC
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("load art post media: %w", err)
	}
	defer rows.Close()

	var media []art.MediaAsset
	for rows.Next() {
		var m art.MediaAsset
		if err := rows.Scan(&m.ID, &m.URL, &m.MimeType, &m.Width, &m.Height, &m.SortOrder); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("load art post media: %w", err)
	}
	if media == nil {
		media = []art.MediaAsset{}
	}
	return media, nil
}

func scanArtPost(row scannable) (art.ArtPost, error) {
	var post art.ArtPost
	var status string
	err := row.Scan(
		&post.ID, &post.ArtistID, &post.Title, &post.Description, &post.Medium,
		&post.Year, &post.Dimensions, &post.City, &post.Style, &post.Residency, &post.Exhibition, &post.FeaturedAcquisition, &post.Palette,
		&status, &post.PublishedAt, &post.CreatedAt, &post.UpdatedAt,
	)
	if err != nil {
		return art.ArtPost{}, err
	}
	post.Status = art.ArtStatus(status)
	return post, nil
}

func coalescePalette(palette []string) []string {
	if palette == nil {
		return []string{}
	}
	return palette
}
