package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
)

type InviteHandler struct {
	invites    inbound.InviteService
	cookieName string
	webAppURL  string
	now        func() time.Time
}

func NewInviteHandler(invites inbound.InviteService, cookieName, webAppURL string) *InviteHandler {
	return &InviteHandler{
		invites:    invites,
		cookieName: cookieName,
		webAppURL:  strings.TrimRight(webAppURL, "/"),
		now:        func() time.Time { return time.Now().UTC() },
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
	status, ok := onboarding.ParseInviteStatusFilter(c.Query("status"))
	if !ok {
		writeError(c, fmt.Errorf("%w: status must be live, claimed, revoked, expired or all", apperrors.ErrValidation))
		return
	}
	invites, total, err := h.invites.ListInvites(c.Request.Context(), onboarding.InviteListFilter{
		Status: status, Limit: queryLimit(c, 50), Offset: queryOffset(c),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("X-Total-Count", strconv.Itoa(total))
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
		writeInviteError(c, err)
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
		writeInviteError(c, err)
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

// Register creates the invitee's account from the invite and signs them in. The account's email is
// read from the invite on the server, so the request carries only the token and a password and can
// never be used to create a verified account for an address the requester merely typed.
func (h *InviteHandler) Register(c *gin.Context) {
	var req dto.InviteRegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	registered, err := h.invites.RegisterWithInvite(c.Request.Context(), req.Token, req.Password)
	if err != nil {
		writeError(c, err)
		return
	}
	writeAuthCookie(c, h.cookieName, registered.AccessToken)
	c.JSON(http.StatusCreated, toUserResponse(registered.User))
}

// writeInviteError is writeError for the admin-only create and resend calls, which also tell the
// admin why an invite conflicts ("already has an account", "handle already taken"). writeError
// hides conflict text because on public endpoints it would reveal which emails are registered; an
// admin already sees that. Only text the use case wrote itself (the sentinel followed by ": ") is
// shown, never a wrapped repository error.
func writeInviteError(c *gin.Context, err error) {
	if errors.Is(err, apperrors.ErrConflict) {
		if reason, ok := strings.CutPrefix(err.Error(), apperrors.ErrConflict.Error()+": "); ok {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: reason})
			return
		}
	}
	writeError(c, err)
}
