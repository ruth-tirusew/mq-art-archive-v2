package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
	authuc "github.com/mq/api/internal/usecase/auth"
)

// verifyingUserRepo is a user repo that can also mark an email verified, like the Postgres one.
type verifyingUserRepo struct {
	*stubUserRepo
	verified map[uuid.UUID]time.Time
}

func (r *verifyingUserRepo) MarkEmailVerified(_ context.Context, id uuid.UUID, at time.Time) error {
	r.verified[id] = at
	return nil
}

func newVerifyingRepo() *verifyingUserRepo {
	return &verifyingUserRepo{
		stubUserRepo: &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		verified:     map[uuid.UUID]time.Time{},
	}
}

func newInviteAuth(users any, mailer *stubMailer, verification *stubVerificationRepo) interface {
	RegisterInvited(ctx context.Context, email, password string) (string, *identity.User, error)
	Register(ctx context.Context, email, password string) (string, *identity.User, error)
} {
	switch u := users.(type) {
	case *verifyingUserRepo:
		return authuc.NewService(u, &stubOAuthAccountRepo{}, stubTokenIssuer{}, stubPasswordHasher{}, nil, nil, nil,
			mailer, nil, nil, "http://localhost:5173", verification)
	case *stubUserRepo:
		return authuc.NewService(u, &stubOAuthAccountRepo{}, stubTokenIssuer{}, stubPasswordHasher{}, nil, nil, nil,
			mailer, nil, nil, "http://localhost:5173", verification)
	}
	panic("unsupported repo")
}

func TestService_RegisterInvited_createsAVerifiedAccountWithoutAVerificationEmail(t *testing.T) {
	users := newVerifyingRepo()
	mailer, verification := &stubMailer{}, &stubVerificationRepo{}
	svc := newInviteAuth(users, mailer, verification)

	token, user, err := svc.RegisterInvited(t.Context(), "  Ada@Example.com ", "password1")
	assist.NoError(t, err)
	assist.Equal(t, "access-token", token)
	assist.Equal(t, "ada@example.com", user.Email)
	assist.Equal(t, identity.RolePublic, user.Role) // an invite grants no role; approval does
	assist.NotNil(t, user.EmailVerifiedAt)
	if _, ok := users.verified[user.ID]; !ok {
		t.Error("the account was not marked verified in the store")
	}
	assist.Equal(t, "hash:password1", users.passwordHashes["ada@example.com"])
	assist.Equal(t, 0, verification.created) // the invite link already proved the inbox
	assist.Equal(t, 0, mailer.sent)
}

func TestService_Register_stillLeavesTheAccountUnverified(t *testing.T) {
	users := newVerifyingRepo()
	mailer, verification := &stubMailer{}, &stubVerificationRepo{}
	svc := newInviteAuth(users, mailer, verification)

	_, user, err := svc.Register(t.Context(), "bez@example.com", "password1")
	assist.NoError(t, err)
	if user.EmailVerifiedAt != nil || len(users.verified) != 0 {
		t.Error("ordinary registration must not verify the email")
	}
	assist.Equal(t, 1, mailer.sent)
}

func TestService_RegisterInvited_rejectsExistingAccountsAndWeakPasswords(t *testing.T) {
	users := newVerifyingRepo()
	existing := identity.User{ID: uuid.New(), Email: "taken@example.com"}
	users.byID[existing.ID] = existing
	users.byEmail[existing.Email] = existing
	svc := newInviteAuth(users, &stubMailer{}, &stubVerificationRepo{})

	_, _, err := svc.RegisterInvited(t.Context(), "taken@example.com", "password1")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	if len(users.verified) != 0 {
		t.Error("an existing account must never be marked verified through an invite")
	}

	_, _, err = svc.RegisterInvited(t.Context(), "new@example.com", "short")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	if _, exists := users.byEmail["new@example.com"]; exists {
		t.Error("a rejected password must not leave an account behind")
	}
}

func TestService_RegisterInvited_refusesRatherThanCreateAnUnverifiedAccount(t *testing.T) {
	// A repo that can't record verification would produce an account that can't claim the invite.
	plain := &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}}
	svc := newInviteAuth(plain, &stubMailer{}, &stubVerificationRepo{})

	_, _, err := svc.RegisterInvited(t.Context(), "new@example.com", "password1")
	assist.ErrorIs(t, err, apperrors.ErrNotImplemented)
	if len(plain.byEmail) != 0 {
		t.Error("no account should have been created")
	}
}
