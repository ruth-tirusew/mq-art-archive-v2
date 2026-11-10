package questionnaire

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

// userRepoStub satisfies outbound.UserRepository; only GetByID is used by the nudge email.
type userRepoStub struct{}

func (userRepoStub) GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	return &identity.User{ID: id, Email: id.String() + "@example.com", DisplayName: "Artist"}, nil
}
func (userRepoStub) GetByEmail(ctx context.Context, e string) (*identity.User, error) {
	return nil, nil
}
func (userRepoStub) GetAuthByEmail(ctx context.Context, e string) (*identity.User, *string, error) {
	return nil, nil, nil
}
func (userRepoStub) GetAuthByID(ctx context.Context, id uuid.UUID) (*identity.User, *string, error) {
	return nil, nil, nil
}
func (userRepoStub) Create(ctx context.Context, u identity.User) error { return nil }
func (userRepoStub) CreateWithPassword(ctx context.Context, u identity.User, h string) error {
	return nil
}
func (userRepoStub) UpdatePassword(ctx context.Context, id uuid.UUID, h string) error    { return nil }
func (userRepoStub) UpdateEmail(ctx context.Context, id uuid.UUID, e string) error       { return nil }
func (userRepoStub) UpdateProfile(ctx context.Context, id uuid.UUID, n, a string) error  { return nil }
func (userRepoStub) UpdateRole(ctx context.Context, id uuid.UUID, r identity.Role) error { return nil }
func (userRepoStub) PromoteToArtist(ctx context.Context, id uuid.UUID) error             { return nil }
func (userRepoStub) PromoteToInstitution(ctx context.Context, id uuid.UUID) error        { return nil }
