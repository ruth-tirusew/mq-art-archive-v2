package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
)

// StaffInviteHandler invites moderators by email. List, Create, Resend and Revoke are admin-only (the router
// enforces it); Preview and Accept are public, because the invitee has no account yet.
type StaffInviteHandler struct {
	invites    inbound.StaffInviteService
	cookieName string
	now        func() time.Time
}

func NewStaffInviteHandler(invites inbound.StaffInviteService, cookieName string) *StaffInviteHandler {
	return &StaffInviteHandler{invites: invites, cookieName: cookieName, now: func() time.Time { return time.Now().UTC() }}
}

func (h *StaffInviteHandler) List(c *gin.Context) {
	invites, total, err := h.invites.ListInvites(c.Request.Context(), queryLimit(c, 50), queryOffset(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.JSON(http.StatusOK, dto.ToStaffInviteResponses(invites, h.now()))
}

func (h *StaffInviteHandler) Create(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.CreateStaffInviteRequest
	if !bindJSON(c, &req) {
		return
	}
	role := identity.Role(req.Role)
	if role == "" {
		role = identity.RoleModerator
	}
	out, err := h.invites.CreateInvite(c.Request.Context(), adminID, inbound.CreateStaffInviteInput{
		Email: req.Email, Role: role, TTL: time.Duration(req.TTLDays) * 24 * time.Hour,
	})
	if err != nil {
		writeInviteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.created(out))
}

func (h *StaffInviteHandler) Resend(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	out, err := h.invites.ResendInvite(c.Request.Context(), adminID, id)
	if err != nil {
		writeInviteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.created(out))
}

func (h *StaffInviteHandler) Revoke(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.invites.RevokeInvite(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *StaffInviteHandler) Preview(c *gin.Context) {
	var req dto.StaffInvitePreviewRequest
	if !bindJSON(c, &req) {
		return
	}
	preview, err := h.invites.PreviewInvite(c.Request.Context(), req.Token)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToStaffInvitePreviewResponse(*preview))
}

func (h *StaffInviteHandler) Accept(c *gin.Context) {
	var req dto.StaffInviteAcceptRequest
	if !bindJSON(c, &req) {
		return
	}
	token, user, err := h.invites.AcceptInvite(c.Request.Context(), req.Token, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	writeAuthCookie(c, h.cookieName, token)
	c.JSON(http.StatusCreated, toUserResponse(user))
}

func (h *StaffInviteHandler) created(out *inbound.StaffInviteWithToken) dto.StaffInviteCreatedResponse {
	return dto.StaffInviteCreatedResponse{Invite: dto.ToStaffInviteResponse(out.Invite, h.now()), Link: out.Link}
}
