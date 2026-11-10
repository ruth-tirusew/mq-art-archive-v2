package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
)

type CreateStaffInviteRequest struct {
	Email string `json:"email" binding:"required,max=254"`
	// Role defaults to moderator, the only role that can be invited by email.
	Role    string `json:"role"`
	TTLDays int    `json:"ttl_days"`
}

type StaffInviteResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	ExpiresAt time.Time  `json:"expires_at"`
	ClaimedAt *time.Time `json:"claimed_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// StaffInviteCreatedResponse adds the one-time link, shown to the admin once. The token is never listed.
type StaffInviteCreatedResponse struct {
	Invite StaffInviteResponse `json:"invite"`
	Link   string              `json:"link"`
}

type StaffInvitePreviewRequest struct {
	Token string `json:"token" binding:"required,max=200"`
}

type StaffInvitePreviewResponse struct {
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}

type StaffInviteAcceptRequest struct {
	Token    string `json:"token" binding:"required,max=200"`
	Password string `json:"password" binding:"required,max=200"`
}

func ToStaffInviteResponse(inv identity.StaffInvite, now time.Time) StaffInviteResponse {
	return StaffInviteResponse{
		ID: inv.ID, Email: inv.Email, Role: string(inv.Role), Status: string(inv.Status(now)),
		ExpiresAt: inv.ExpiresAt, ClaimedAt: inv.ClaimedAt, CreatedAt: inv.CreatedAt,
	}
}

func ToStaffInviteResponses(invs []identity.StaffInvite, now time.Time) []StaffInviteResponse {
	out := make([]StaffInviteResponse, len(invs))
	for i, inv := range invs {
		out[i] = ToStaffInviteResponse(inv, now)
	}
	return out
}

func ToStaffInvitePreviewResponse(p inbound.StaffInvitePreview) StaffInvitePreviewResponse {
	return StaffInvitePreviewResponse{Email: p.Email, Role: string(p.Role), Status: string(p.Status), ExpiresAt: p.ExpiresAt}
}
