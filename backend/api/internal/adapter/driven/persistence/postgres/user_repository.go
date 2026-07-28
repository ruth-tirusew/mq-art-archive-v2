package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
)

type UserRepository struct {
	pool *Pool
}

func NewUserRepository(pool *Pool) outbound.UserRepository {
	return &UserRepository{pool: pool}
}

const userColumns = `id, email, role, COALESCE(display_name, ''), COALESCE(avatar_url, ''), email_verified_at, created_at, updated_at`

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE id = $1
	`, id)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*identity.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`
		FROM users
		WHERE email = $1
	`, email)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) GetAuthByEmail(ctx context.Context, email string) (*identity.User, *string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`, password_hash
		FROM users
		WHERE email = $1
	`, email)
	return scanAuthUser(row)
}

func (r *UserRepository) GetAuthByID(ctx context.Context, id uuid.UUID) (*identity.User, *string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`, password_hash
		FROM users
		WHERE id = $1
	`, id)
	return scanAuthUser(row)
}

func (r *UserRepository) List(ctx context.Context, role *identity.Role, limit, offset int) ([]identity.User, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	where := ""
	args := []any{}
	if role != nil {
		where = " WHERE role=$1"
		args = append(args, string(*role))
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, "SELECT "+userColumns+" FROM users"+where+
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	users := []identity.User{}
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, rows.Err()
}

func (r *UserRepository) CountByRole(ctx context.Context, role identity.Role) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE role=$1`, string(role)).Scan(&count)
	return count, err
}

func scanAuthUser(row scannable) (*identity.User, *string, error) {
	var user identity.User
	var role string
	var passwordHash *string
	err := row.Scan(
		&user.ID, &user.Email, &role, &user.DisplayName, &user.AvatarURL, &user.EmailVerifiedAt,
		&user.CreatedAt, &user.UpdatedAt, &passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("get auth user: %w", err)
	}
	user.Role = identity.Role(role)
	user.HasPassword = passwordHash != nil && *passwordHash != ""
	return &user, passwordHash, nil
}

func (r *UserRepository) Create(ctx context.Context, user identity.User) error {
	return r.insertUser(ctx, user, nil)
}

func (r *UserRepository) CreateWithPassword(ctx context.Context, user identity.User, passwordHash string) error {
	return r.insertUser(ctx, user, &passwordHash)
}

func (r *UserRepository) insertUser(ctx context.Context, user identity.User, passwordHash *string) error {
	now := time.Now().UTC()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, email, role, display_name, avatar_url, created_at, updated_at, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, string(user.Role), emptyToNil(user.DisplayName), emptyToNil(user.AvatarURL),
		user.CreatedAt, user.UpdatedAt, passwordHash)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`, id, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users
		SET email = $2, updated_at = NOW()
		WHERE id = $1
	`, id, email)
	if err != nil {
		return fmt.Errorf("update email: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uuid.UUID, at time.Time) error {
	tag, err := r.pool.Exec(ctx, `UPDATE users SET email_verified_at=$2, updated_at=$2 WHERE id=$1`, id, at)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, displayName, avatarURL string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users
		SET display_name = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $1
	`, id, emptyToNil(displayName), emptyToNil(avatarURL))
	if err != nil {
		return fmt.Errorf("update profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role identity.Role) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE users
		SET role = $2, updated_at = NOW()
		WHERE id = $1
	`, id, string(role))
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *UserRepository) PromoteToArtist(ctx context.Context, id uuid.UUID) error {
	return r.UpdateRole(ctx, id, identity.RoleArtist)
}

func (r *UserRepository) PromoteToInstitution(ctx context.Context, id uuid.UUID) error {
	return r.UpdateRole(ctx, id, identity.RoleInstitution)
}

func scanUser(row scannable) (identity.User, error) {
	var user identity.User
	var role string
	err := row.Scan(&user.ID, &user.Email, &role, &user.DisplayName, &user.AvatarURL, &user.EmailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return identity.User{}, err
	}
	user.Role = identity.Role(role)
	return user, nil
}

func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
