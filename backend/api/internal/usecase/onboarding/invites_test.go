package onboarding

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/testutil/assist"
)

// fakeInvites is an in-memory InviteRepository with the same live-handle and single-claim rules
// as the Postgres one.
type fakeInvites struct{ byID map[uuid.UUID]*domain.Invite }

func newFakeInvites() *fakeInvites { return &fakeInvites{byID: map[uuid.UUID]*domain.Invite{}} }

func (f *fakeInvites) live(i *domain.Invite, now time.Time) bool {
	return i.Status(now) == domain.InviteLive
}

func (f *fakeInvites) Create(ctx context.Context, inv domain.Invite) (*domain.Invite, error) {
	for _, o := range f.byID {
		if o.TokenHash == inv.TokenHash {
			return nil, apperrors.ErrConflict
		}
		if inv.RequestedHandle != "" && strings.EqualFold(o.RequestedHandle, inv.RequestedHandle) && o.ClaimedAt == nil && o.RevokedAt == nil {
			if !o.ExpiresAt.After(inv.CreatedAt) {
				t := inv.CreatedAt
				o.RevokedAt = &t
				continue
			}
			return nil, apperrors.ErrConflict
		}
	}
	f.byID[inv.ID] = &inv
	return &inv, nil
}
func (f *fakeInvites) GetByID(ctx context.Context, id uuid.UUID) (*domain.Invite, error) {
	if i, ok := f.byID[id]; ok {
		c := *i
		return &c, nil
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeInvites) GetByTokenHash(ctx context.Context, h string) (*domain.Invite, error) {
	for _, i := range f.byID {
		if i.TokenHash == h {
			c := *i
			return &c, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeInvites) GetLiveByHandle(ctx context.Context, handle string, now time.Time) (*domain.Invite, error) {
	for _, i := range f.byID {
		if i.RequestedHandle != "" && strings.EqualFold(i.RequestedHandle, handle) && f.live(i, now) {
			c := *i
			return &c, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeInvites) List(ctx context.Context) ([]domain.Invite, error) {
	out := []domain.Invite{}
	for _, i := range f.byID {
		out = append(out, *i)
	}
	return out, nil
}
func (f *fakeInvites) Claim(ctx context.Context, id, userID uuid.UUID, now time.Time) (*domain.Invite, error) {
	i, ok := f.byID[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	if !f.live(i, now) {
		return nil, apperrors.ErrConflict
	}
	i.ClaimedBy, i.ClaimedAt = &userID, &now
	c := *i
	return &c, nil
}
func (f *fakeInvites) Revoke(ctx context.Context, id uuid.UUID, now time.Time) error {
	i, ok := f.byID[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	if i.ClaimedAt != nil {
		return apperrors.ErrConflict
	}
	if i.RevokedAt == nil {
		i.RevokedAt = &now
	}
	return nil
}

type userStub struct {
	mockUserRepo
	user *identity.User
}

func (u *userStub) GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	if u.user != nil && u.user.ID == id {
		c := *u.user
		return &c, nil
	}
	return nil, apperrors.ErrNotFound
}

type gateStub struct {
	rev *uuid.UUID
	err error
}

func (g gateStub) CheckComplete(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	return g.rev, g.err
}

type inviteEnv struct {
	svc     *Service
	invites *fakeInvites
	saved   []domain.OnboardingApplication
	user    *identity.User
	now     time.Time
}

func newInviteEnv(t *testing.T, gate inbound.ApplicationQuestionnaireGate) *inviteEnv {
	t.Helper()
	verified := time.Now().Add(-time.Hour)
	env := &inviteEnv{
		invites: newFakeInvites(),
		user:    &identity.User{ID: uuid.New(), Email: "Ada@Example.com", DisplayName: "Ada", EmailVerifiedAt: &verified},
		now:     time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC),
	}
	apps := &mockOnboardingRepo{
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			env.saved = append(env.saved, app)
			return &app, nil
		},
	}
	opts := []Option{WithInvites(env.invites)}
	if gate != nil {
		opts = append(opts, WithQuestionnaire(gate))
	}
	env.svc = NewService(apps, &userStub{user: env.user}, &mockProfileRepo{}, &mockInstitutionRepo{}, opts...)
	env.svc.now = func() time.Time { return env.now }
	return env
}

func (e *inviteEnv) invite(t *testing.T, handle string) *inbound.InviteWithToken {
	t.Helper()
	got, err := e.svc.CreateInvite(context.Background(), uuid.New(), inbound.CreateInviteInput{
		Email: "ada@example.com", DisplayName: "Ada Lovelace", Handle: handle,
	})
	assist.NoError(t, err)
	return got
}

func TestCreateInvite_storesOnlyHashAndReservesHandle(t *testing.T) {
	env := newInviteEnv(t, nil)
	got := env.invite(t, "ada_1")

	assist.NotEqual(t, got.Token, got.Invite.TokenHash)
	assist.Equal(t, domain.HashInviteToken(got.Token), got.Invite.TokenHash)
	if len(got.Token) < 40 {
		t.Errorf("token looks too short to be 256 bits: %q", got.Token)
	}
	assist.Equal(t, env.now.Add(14*24*time.Hour), got.Invite.ExpiresAt)

	ok, err := env.svc.CheckHandleAvailable(context.Background(), "ada_1")
	assist.NoError(t, err)
	assist.Equal(t, false, ok) // reserved by the live invite

	_, err = env.svc.CreateInvite(context.Background(), uuid.New(), inbound.CreateInviteInput{Email: "b@example.com", Handle: "ADA_1"})
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestCreateInvite_validation(t *testing.T) {
	env := newInviteEnv(t, nil)
	for name, in := range map[string]inbound.CreateInviteInput{
		"bad email":     {Email: "not-an-email"},
		"display email": {Email: "Ada <ada@example.com>"},
		"bad handle":    {Email: "a@example.com", Handle: "No Spaces"},
		"ttl too short": {Email: "a@example.com", TTL: time.Minute},
		"ttl too long":  {Email: "a@example.com", TTL: 90 * 24 * time.Hour},
		"negative ttl":  {Email: "a@example.com", TTL: -time.Hour},
	} {
		_, err := env.svc.CreateInvite(context.Background(), uuid.New(), in)
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		assist.ErrorIs(t, err, apperrors.ErrValidation)
	}
}

func TestSubmitWithInvite_happyPath(t *testing.T) {
	rev := uuid.New()
	env := newInviteEnv(t, gateStub{rev: &rev})
	inv := env.invite(t, "ada_1")

	app, err := env.svc.SubmitWithInvite(context.Background(), env.user.ID, inv.Token, "", "", "")
	assist.NoError(t, err)
	assist.Equal(t, "Ada Lovelace", app.DisplayName)          // defaulted from the invite
	assist.Equal(t, "ada_1", app.RequestedHandle)             // its own reservation doesn't block it
	assist.Equal(t, domain.ApprovalStatusPending, app.Status) // an invite never skips review
	assist.NotNil(t, app.InviteID)
	assist.Equal(t, inv.Invite.ID, *app.InviteID)
	assist.NotNil(t, app.QuestionnaireRevisionID)
	assist.Equal(t, rev, *app.QuestionnaireRevisionID)

	got, _ := env.invites.GetByID(context.Background(), inv.Invite.ID)
	assist.Equal(t, domain.InviteClaimed, got.Status(env.now))

	// Single use.
	_, err = env.svc.SubmitWithInvite(context.Background(), env.user.ID, inv.Token, "", "", "")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestSubmitWithInvite_incompleteQuestionnaireDoesNotBurnInvite(t *testing.T) {
	env := newInviteEnv(t, gateStub{err: fmt.Errorf("%w: answer the required questions", apperrors.ErrValidation)})
	inv := env.invite(t, "ada_1")

	_, err := env.svc.SubmitWithInvite(context.Background(), env.user.ID, inv.Token, "", "", "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
	assist.Len(t, 0, len(env.saved))

	got, _ := env.invites.GetByID(context.Background(), inv.Invite.ID)
	assist.Equal(t, domain.InviteLive, got.Status(env.now)) // still usable once they finish
}

func TestSubmitWithInvite_rejections(t *testing.T) {
	ctx := context.Background()

	t.Run("unverified email", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		env.user.EmailVerifiedAt = nil
		inv := env.invite(t, "")
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, inv.Token, "Ada", "ada_1", "")
		assist.ErrorIs(t, err, apperrors.ErrForbidden)
	})
	t.Run("different email", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		env.user.Email = "someone.else@example.com"
		inv := env.invite(t, "")
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, inv.Token, "Ada", "ada_1", "")
		assist.ErrorIs(t, err, apperrors.ErrForbidden)
		got, _ := env.invites.GetByID(ctx, inv.Invite.ID)
		assist.Equal(t, domain.InviteLive, got.Status(env.now))
	})
	t.Run("expired", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		inv := env.invite(t, "")
		env.now = env.now.Add(15 * 24 * time.Hour)
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, inv.Token, "Ada", "ada_1", "")
		assist.ErrorIs(t, err, apperrors.ErrConflict)
	})
	t.Run("revoked", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		inv := env.invite(t, "")
		assist.NoError(t, env.svc.RevokeInvite(ctx, inv.Invite.ID))
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, inv.Token, "Ada", "ada_1", "")
		assist.ErrorIs(t, err, apperrors.ErrConflict)
	})
	t.Run("unknown token", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, "nope", "Ada", "ada_1", "")
		assist.ErrorIs(t, err, apperrors.ErrNotFound)
	})
	t.Run("handle needed when invite has none", func(t *testing.T) {
		env := newInviteEnv(t, nil)
		inv := env.invite(t, "")
		_, err := env.svc.SubmitWithInvite(ctx, env.user.ID, inv.Token, "Ada", "", "")
		assist.ErrorIs(t, err, apperrors.ErrValidation)
	})
}

func TestResendInvite_replacesOldLink(t *testing.T) {
	env := newInviteEnv(t, nil)
	ctx := context.Background()
	first := env.invite(t, "ada_1")

	second, err := env.svc.ResendInvite(ctx, uuid.New(), first.Invite.ID)
	assist.NoError(t, err)
	assist.NotEqual(t, first.Token, second.Token)
	assist.Equal(t, "ada_1", second.Invite.RequestedHandle) // handle carries over

	_, err = env.svc.SubmitWithInvite(ctx, env.user.ID, first.Token, "", "", "")
	assist.ErrorIs(t, err, apperrors.ErrConflict) // old link is dead
	_, err = env.svc.SubmitWithInvite(ctx, env.user.ID, second.Token, "", "", "")
	assist.NoError(t, err)

	_, err = env.svc.ResendInvite(ctx, uuid.New(), second.Invite.ID)
	assist.ErrorIs(t, err, apperrors.ErrConflict) // already used
}

func TestPreviewInvite(t *testing.T) {
	env := newInviteEnv(t, nil)
	inv := env.invite(t, "ada_1")
	p, err := env.svc.PreviewInvite(context.Background(), inv.Token)
	assist.NoError(t, err)
	assist.Equal(t, "ada@example.com", p.Email)
	assist.Equal(t, domain.InviteLive, p.Status)
	_, err = env.svc.PreviewInvite(context.Background(), "unknown")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestInvites_notConfigured(t *testing.T) {
	svc := NewService(&mockOnboardingRepo{}, &mockUserRepo{}, &mockProfileRepo{}, &mockInstitutionRepo{})
	_, err := svc.CreateInvite(context.Background(), uuid.New(), inbound.CreateInviteInput{Email: "a@example.com"})
	assist.ErrorIs(t, err, apperrors.ErrNotImplemented)
}

func TestSubmit_requiresCompleteQuestionnaireForArtistsOnly(t *testing.T) {
	gateErr := fmt.Errorf("%w: incomplete", apperrors.ErrValidation)
	env := newInviteEnv(t, gateStub{err: gateErr})

	_, err := env.svc.SubmitWithHandle(context.Background(), uuid.New(), domain.ApplicantTypeArtist, "Studio", "studio_x", "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)

	_, err = env.svc.SubmitWithHandle(context.Background(), uuid.New(), domain.ApplicantTypeInstitution, "Museum", "", "")
	assist.NoError(t, err) // institutions don't answer the artist questionnaire
}
