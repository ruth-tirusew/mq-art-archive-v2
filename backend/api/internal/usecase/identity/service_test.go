package identity

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/identity"
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

func (m *mockUserRepo) GetAuthByEmail(_ context.Context, _ string) (*domain.User, *string, error) {
	return nil, nil, apperrors.ErrNotFound
}

func (m *mockUserRepo) GetAuthByID(_ context.Context, _ uuid.UUID) (*domain.User, *string, error) {
	return nil, nil, apperrors.ErrNotFound
}

func (m *mockUserRepo) Create(_ context.Context, _ domain.User) error {
	return nil
}

func (m *mockUserRepo) CreateWithPassword(_ context.Context, _ domain.User, _ string) error {
	return nil
}
func (m *mockUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockUserRepo) UpdateEmail(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (m *mockUserRepo) UpdateProfile(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}
func (m *mockUserRepo) UpdateRole(_ context.Context, _ uuid.UUID, _ domain.Role) error {
	return nil
}
func (m *mockUserRepo) PromoteToArtist(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (m *mockUserRepo) PromoteToInstitution(_ context.Context, _ uuid.UUID) error {
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
