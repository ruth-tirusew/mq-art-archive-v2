package inbound

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/onboarding"
)

type OnboardingService interface {
	ListPending(ctx context.Context) ([]onboarding.OnboardingApplication, error)
	GetByID(ctx context.Context, id uuid.UUID) (*onboarding.OnboardingApplication, error)
	GetMyApplication(ctx context.Context, applicantID uuid.UUID) (*onboarding.OnboardingApplication, error)
	Submit(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, notes string) (*onboarding.OnboardingApplication, error)
	SubmitWithHandle(ctx context.Context, applicantID uuid.UUID, applicantType onboarding.ApplicantType, displayName, requestedHandle, notes string) (*onboarding.OnboardingApplication, error)
	CheckHandleAvailable(ctx context.Context, handle string) (bool, error)
	Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status onboarding.ApprovalStatus, notes string) (*onboarding.OnboardingApplication, error)
}

type CreateInviteInput struct {
	Email       string
	DisplayName string
	Handle      string        // optional; reserved for the invitee while the invite is live
	TTL         time.Duration // 0 means the default (14 days)
}

// InviteWithToken carries the raw token, which exists only at creation or resend: it is not stored
// and cannot be recovered later, so the admin UI should show the link once.
type InviteWithToken struct {
	Invite onboarding.Invite
	Token  string
}

// InvitePreview is what the claim page shows before the invitee signs in.
type InvitePreview struct {
	Email       string
	DisplayName string
	Handle      string
	Status      onboarding.InviteStatus
}

// RegisteredInvitee is a freshly created invitee account and its session token.
type RegisteredInvitee struct {
	AccessToken string
	User        *identity.User
}

type InviteService interface {
	CreateInvite(ctx context.Context, adminID uuid.UUID, in CreateInviteInput) (*InviteWithToken, error)
	// ResendInvite revokes the invite and issues a fresh link with the same details.
	ResendInvite(ctx context.Context, adminID, inviteID uuid.UUID) (*InviteWithToken, error)
	RevokeInvite(ctx context.Context, inviteID uuid.UUID) error
	ListInvites(ctx context.Context) ([]onboarding.Invite, error)
	// PreviewInvite looks an invite up by its raw token. ErrNotFound for an unknown token.
	PreviewInvite(ctx context.Context, rawToken string) (*InvitePreview, error)
	// RegisterWithInvite creates the invitee's verified account and signs them in. The email is
	// read from the live invite, never taken from the request, so it cannot be used to create a
	// verified account for an address the requester merely typed. ErrNotFound for an unknown token,
	// ErrConflict if the invite was used, revoked or expired, or the address already has an account.
	RegisterWithInvite(ctx context.Context, rawToken, password string) (*RegisteredInvitee, error)
	// SubmitWithInvite claims the invite and submits the application in one step. The signed-in
	// user's verified email must match the invite's.
	SubmitWithInvite(ctx context.Context, applicantID uuid.UUID, rawToken, displayName, requestedHandle, notes string) (*onboarding.OnboardingApplication, error)
}
