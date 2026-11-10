package auth_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	authuc "github.com/mq/api/internal/usecase/auth"
)

// fakeStaffInvites is an in-memory StaffInviteRepository with the same rules as the Postgres one: one live
// invite per address, a single-use claim, and an account that is created only if the invite is still live.
type fakeStaffInvites struct {
	invites  map[uuid.UUID]*identity.StaffInvite
	accounts map[string]identity.User // by lower-case email
	now      func() time.Time
}

func newFakeStaffInvites(now func() time.Time) *fakeStaffInvites {
	return &fakeStaffInvites{invites: map[uuid.UUID]*identity.StaffInvite{}, accounts: map[string]identity.User{}, now: now}
}

func (f *fakeStaffInvites) Create(_ context.Context, inv identity.StaffInvite) (*identity.StaffInvite, error) {
	for _, o := range f.invites {
		if strings.EqualFold(o.Email, inv.Email) && o.ClaimedAt == nil && o.RevokedAt == nil {
			t := inv.CreatedAt
			o.RevokedAt = &t // a new invite replaces the old live one
		}
	}
	f.invites[inv.ID] = &inv
	return &inv, nil
}
func (f *fakeStaffInvites) GetByID(_ context.Context, id uuid.UUID) (*identity.StaffInvite, error) {
	if i, ok := f.invites[id]; ok {
		c := *i
		return &c, nil
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeStaffInvites) GetByTokenHash(_ context.Context, h string) (*identity.StaffInvite, error) {
	for _, i := range f.invites {
		if i.TokenHash == h {
			c := *i
			return &c, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeStaffInvites) List(context.Context, int, int) ([]identity.StaffInvite, int, error) {
	out := []identity.StaffInvite{}
	for _, i := range f.invites {
		out = append(out, *i)
	}
	return out, len(out), nil
}
func (f *fakeStaffInvites) Accept(_ context.Context, id uuid.UUID, user identity.User, _ string, now time.Time) (*identity.User, error) {
	inv, ok := f.invites[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	if _, exists := f.accounts[strings.ToLower(user.Email)]; exists || inv.Status(now) != identity.StaffInviteLive {
		return nil, apperrors.ErrConflict
	}
	user.ID = uuid.New()
	f.accounts[strings.ToLower(user.Email)] = user
	inv.ClaimedAt, inv.ClaimedBy = &now, &user.ID
	return &user, nil
}
func (f *fakeStaffInvites) Revoke(_ context.Context, id uuid.UUID, now time.Time) error {
	inv, ok := f.invites[id]
	if !ok {
		return apperrors.ErrNotFound
	}
	if inv.ClaimedAt != nil {
		return apperrors.ErrConflict
	}
	inv.RevokedAt = &now
	return nil
}

type staffEnv struct {
	svc    inbound.StaffInviteService
	repo   *fakeStaffInvites
	users  *stubUserRepo
	mailer *stubMailer
	admin  uuid.UUID
}

func newStaffEnv(t *testing.T) *staffEnv {
	t.Helper()
	e := &staffEnv{
		users:  &stubUserRepo{byID: map[uuid.UUID]identity.User{}, byEmail: map[string]identity.User{}},
		mailer: &stubMailer{},
		admin:  uuid.New(),
	}
	e.repo = newFakeStaffInvites(time.Now)
	e.svc = authuc.NewStaffInviteService(e.users, stubPasswordHasher{}, stubTokenIssuer{}, e.repo, e.mailer, "https://admin.artiv.test/")
	return e
}

func (e *staffEnv) invite(t *testing.T, email string) *inbound.StaffInviteWithToken {
	t.Helper()
	out, err := e.svc.CreateInvite(context.Background(), e.admin, inbound.CreateStaffInviteInput{Email: email, Role: identity.RoleModerator})
	if err != nil {
		t.Fatalf("create invite for %s: %v", email, err)
	}
	return out
}

func TestStaffInvite_createsALinkAndOnlyStoresItsHash(t *testing.T) {
	e := newStaffEnv(t)
	out := e.invite(t, "  New.Mod@Example.org ")

	if out.Invite.Email != "new.mod@example.org" || out.Invite.Role != identity.RoleModerator {
		t.Fatalf("address should be normalised and the role fixed, got %+v", out.Invite)
	}
	if !strings.HasPrefix(out.Link, "https://admin.artiv.test/accept-invite?token=") || !strings.Contains(out.Link, out.Token) {
		t.Fatalf("link should point at the admin app and carry the token: %s", out.Link)
	}
	if out.Invite.TokenHash == out.Token || out.Invite.TokenHash != identity.HashStaffInviteToken(out.Token) {
		t.Fatal("only the hash of the token may be stored")
	}
	if len(out.Token) < 40 {
		t.Errorf("token looks too short to be 256 bits: %q", out.Token)
	}
	if got := out.Invite.ExpiresAt.Sub(out.Invite.CreatedAt); got != 7*24*time.Hour {
		t.Errorf("default lifetime should be 7 days, got %s", got)
	}
	if e.mailer.sent != 1 {
		t.Errorf("one email should go out, got %d", e.mailer.sent)
	}
}

func TestStaffInvite_onlyModeratorsCanBeInvitedByEmail(t *testing.T) {
	e := newStaffEnv(t)
	for _, role := range []identity.Role{identity.RoleAdmin, identity.RoleArtist, identity.RolePublic, ""} {
		_, err := e.svc.CreateInvite(context.Background(), e.admin, inbound.CreateStaffInviteInput{Email: "a@example.org", Role: role})
		if err == nil || !strings.Contains(err.Error(), "validation") {
			t.Errorf("role %q must be refused with a validation error, got %v", role, err)
		}
	}
	if len(e.repo.invites) != 0 {
		t.Errorf("nothing may be stored for a refused role, got %d", len(e.repo.invites))
	}
}

func TestStaffInvite_refusesAnAddressThatAlreadyHasAnAccount(t *testing.T) {
	e := newStaffEnv(t)
	e.users.byEmail["taken@example.org"] = identity.User{ID: uuid.New(), Email: "taken@example.org", Role: identity.RoleArtist}

	_, err := e.svc.CreateInvite(context.Background(), e.admin, inbound.CreateStaffInviteInput{Email: "Taken@Example.org", Role: identity.RoleModerator})
	if err == nil || !strings.Contains(err.Error(), "already has an Artiv account") {
		t.Fatalf("want a conflict that says why, got %v", err)
	}
	if len(e.repo.invites) != 0 || e.mailer.sent != 0 {
		t.Errorf("nothing stored and nothing emailed for an existing account (stored=%d sent=%d)", len(e.repo.invites), e.mailer.sent)
	}
}

func TestStaffInvite_lifetimeBounds(t *testing.T) {
	e := newStaffEnv(t)
	for name, ttl := range map[string]time.Duration{"too short": time.Minute, "too long": 31 * 24 * time.Hour, "negative": -time.Hour} {
		if _, err := e.svc.CreateInvite(context.Background(), e.admin, inbound.CreateStaffInviteInput{Email: "a@example.org", Role: identity.RoleModerator, TTL: ttl}); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestStaffInvite_acceptCreatesAVerifiedModeratorAndSignsThemIn(t *testing.T) {
	e := newStaffEnv(t)
	out := e.invite(t, "mod@example.org")

	token, user, err := e.svc.AcceptInvite(context.Background(), out.Token, "a-good-password")
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if token == "" || user == nil {
		t.Fatalf("want a session and a user, got %q %v", token, user)
	}
	if user.Role != identity.RoleModerator || user.Email != "mod@example.org" || user.EmailVerifiedAt == nil || !user.HasPassword {
		t.Errorf("want a verified moderator with a password on the invited address, got %+v", user)
	}

	// The link is single-use.
	if _, _, err := e.svc.AcceptInvite(context.Background(), out.Token, "a-good-password"); err == nil {
		t.Error("a used link must not work a second time")
	}
}

func TestStaffInvite_aTypoInThePasswordDoesNotBurnTheLink(t *testing.T) {
	e := newStaffEnv(t)
	out := e.invite(t, "mod@example.org")

	if _, _, err := e.svc.AcceptInvite(context.Background(), out.Token, "short"); err == nil || !strings.Contains(err.Error(), "validation") {
		t.Fatalf("a weak password is a validation error, got %v", err)
	}
	if _, _, err := e.svc.AcceptInvite(context.Background(), out.Token, "a-good-password"); err != nil {
		t.Errorf("the link must still work after a rejected password: %v", err)
	}
}

func TestStaffInvite_deadAndUnknownLinksCreateNothing(t *testing.T) {
	e := newStaffEnv(t)
	ctx := context.Background()

	if _, _, err := e.svc.AcceptInvite(ctx, "no-such-token", "a-good-password"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("unknown token should be not found, got %v", err)
	}

	revoked := e.invite(t, "revoked@example.org")
	if err := e.svc.RevokeInvite(ctx, revoked.Invite.ID); err != nil {
		t.Fatal(err)
	}
	expired := e.invite(t, "expired@example.org")
	e.repo.invites[expired.Invite.ID].ExpiresAt = time.Now().Add(-time.Minute)

	for name, tok := range map[string]string{"revoked": revoked.Token, "expired": expired.Token} {
		if _, _, err := e.svc.AcceptInvite(ctx, tok, "a-good-password"); err == nil {
			t.Errorf("%s link must be refused", name)
		}
	}
	if len(e.repo.accounts) != 0 {
		t.Errorf("no account may come from a dead link, got %d", len(e.repo.accounts))
	}
}

func TestStaffInvite_previewShowsWhoAndWhetherItStillWorks(t *testing.T) {
	e := newStaffEnv(t)
	out := e.invite(t, "mod@example.org")

	p, err := e.svc.PreviewInvite(context.Background(), out.Token)
	if err != nil {
		t.Fatal(err)
	}
	if p.Email != "mod@example.org" || p.Role != identity.RoleModerator || p.Status != identity.StaffInviteLive {
		t.Errorf("preview = %+v", p)
	}
	if _, err := e.svc.PreviewInvite(context.Background(), "nope"); err == nil {
		t.Error("an unknown token must not preview")
	}
}

func TestStaffInvite_resendReplacesTheOldLinkAndKeepsTheLifetime(t *testing.T) {
	e := newStaffEnv(t)
	first, err := e.svc.CreateInvite(context.Background(), e.admin, inbound.CreateStaffInviteInput{Email: "mod@example.org", Role: identity.RoleModerator, TTL: 3 * 24 * time.Hour})
	if err != nil {
		t.Fatal(err)
	}

	second, err := e.svc.ResendInvite(context.Background(), e.admin, first.Invite.ID)
	if err != nil {
		t.Fatalf("resend: %v", err)
	}
	if second.Token == first.Token {
		t.Fatal("a resend must issue a fresh link")
	}
	if got := second.Invite.ExpiresAt.Sub(second.Invite.CreatedAt); got != 3*24*time.Hour {
		t.Errorf("the original lifetime should be kept, got %s", got)
	}
	if _, _, err := e.svc.AcceptInvite(context.Background(), first.Token, "a-good-password"); err == nil {
		t.Error("the old link must stop working")
	}
	if _, _, err := e.svc.AcceptInvite(context.Background(), second.Token, "a-good-password"); err != nil {
		t.Errorf("the new link must work: %v", err)
	}
}

func TestStaffInvite_aUsedInviteCannotBeResentOrRevoked(t *testing.T) {
	e := newStaffEnv(t)
	out := e.invite(t, "mod@example.org")
	if _, _, err := e.svc.AcceptInvite(context.Background(), out.Token, "a-good-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.svc.ResendInvite(context.Background(), e.admin, out.Invite.ID); err == nil {
		t.Error("resending a used invite must be refused")
	}
	if err := e.svc.RevokeInvite(context.Background(), out.Invite.ID); err == nil {
		t.Error("revoking a used invite must be refused")
	}
}
