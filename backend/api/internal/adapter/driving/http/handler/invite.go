package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/port/inbound"
)

type InviteHandler struct {
	invites   inbound.InviteService
	webAppURL string
	now       func() time.Time
}

func NewInviteHandler(invites inbound.InviteService, webAppURL string) *InviteHandler {
	return &InviteHandler{
		invites:   invites,
		webAppURL: strings.TrimRight(webAppURL, "/"),
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (h *InviteHandler) created(w *inbound.InviteWithToken) dto.InviteCreatedResponse {
	return dto.InviteCreatedResponse{
		Invite: dto.ToInviteResponse(w.Invite, h.now()),
		Token:  w.Token,
		Link:   h.webAppURL + "/apply?invite=" + w.Token,
	}
}

func (h *InviteHandler) List(c *gin.Context) {
	invites, err := h.invites.ListInvites(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToInviteResponses(invites, h.now()))
}

func (h *InviteHandler) Create(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.CreateInviteRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.invites.CreateInvite(c.Request.Context(), adminID, inbound.CreateInviteInput{
		Email: req.Email, DisplayName: req.DisplayName, Handle: req.Handle,
		TTL: time.Duration(req.TTLDays) * 24 * time.Hour,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.created(out))
}

func (h *InviteHandler) Resend(c *gin.Context) {
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
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.created(out))
}

func (h *InviteHandler) Revoke(c *gin.Context) {
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

// Preview lets the claim page show who an invite is for before sign-in. The token travels in the
// POST body, not the URL, so it stays out of access logs and browser history.
func (h *InviteHandler) Preview(c *gin.Context) {
	var req dto.InvitePreviewRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := h.invites.PreviewInvite(c.Request.Context(), req.Token)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.InvitePreviewResponse{
		Email: p.Email, DisplayName: p.DisplayName, Handle: p.Handle, Status: string(p.Status),
	})
}

// Claim claims an invite and submits the application in one step.
func (h *InviteHandler) Claim(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.ClaimInviteRequest
	if !bindJSON(c, &req) {
		return
	}
	app, err := h.invites.SubmitWithInvite(c.Request.Context(), userID, req.Token, req.DisplayName, req.RequestedHandle, req.Notes)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToOnboardingApplicationResponse(*app))
}
