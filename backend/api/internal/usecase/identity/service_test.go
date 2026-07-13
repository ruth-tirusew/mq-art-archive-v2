package identity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/testutil/assist"
)

type mockUserRepo struct {
	getByID func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return m.getByID(ctx, id)
}

func (m *mockUserRepo) GetByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, apperrors.ErrNotFound
}

func (m *mockUserRepo) Create(_ context.Context, _ domain.User) error {
	return nil
}

func TestService_GetUser_delegatesToRepository(t *testing.T) {
	userID := uuid.New()
	expected := &domain.User{ID: userID, Email: "a@example.com", Role: domain.RoleArtist}

	repo := &mockUserRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			assist.Equal(t, userID, id)
			return expected, nil
		},
	}
	svc := NewService(repo)

	got, err := svc.GetUser(context.Background(), userID)
	assist.NoError(t, err)
	assist.Equal(t, expected, got)
}
