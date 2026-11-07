package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/port/inbound"
)

type QuestionnaireHandler struct {
	questionnaire inbound.QuestionnaireService
	admin         inbound.QuestionnaireAdminService
}

func NewQuestionnaireHandler(
	questionnaire inbound.QuestionnaireService,
	admin inbound.QuestionnaireAdminService,
) *QuestionnaireHandler {
	return &QuestionnaireHandler{questionnaire: questionnaire, admin: admin}
}

func currentUser(c *gin.Context) (uuid.UUID, bool) {
	id, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return uuid.Nil, false
	}
	return id, true
}

func bindJSON(c *gin.Context, req any) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return false
	}
	return true
}

func idParam(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuidParam(c, name)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid " + name})
		return uuid.Nil, false
	}
	return id, true
}

// ---- applicant / artist (any signed-in user) ----

func (h *QuestionnaireHandler) GetForm(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	form, err := h.questionnaire.GetForm(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToFormResponse(*form))
}

func (h *QuestionnaireHandler) SaveAnswers(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.SaveAnswersRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.questionnaire.SaveAnswers(c.Request.Context(), userID, dto.ToSaveAnswersInput(req)); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *QuestionnaireHandler) GetNudge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	prompt, err := h.questionnaire.ActiveNudge(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToNudgePromptResponse(*prompt))
}

func (h *QuestionnaireHandler) DismissNudge(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	revisionID, ok := idParam(c, "revisionId")
	if !ok {
		return
	}
	if err := h.questionnaire.DismissNudge(c.Request.Context(), userID, revisionID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- public ----

// ListPublicAnswers serves an artist's visible answers. The profile must be public: a draft,
// pending or taken-down profile answers 404 exactly like an unknown slug, so its answers (which
// may exist, since artists fill the form before approval) are never exposed.
func (h *QuestionnaireHandler) ListPublicAnswers(c *gin.Context) {
	answers, err := h.questionnaire.PublicAnswersBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToPublicAnswerResponses(answers))
}

// ---- admin: editor ----

func (h *QuestionnaireHandler) GetDraft(c *gin.Context) {
	rev, err := h.admin.GetDraft(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponse(*rev))
}

func (h *QuestionnaireHandler) StartDraft(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	rev, err := h.admin.StartDraft(c.Request.Context(), adminID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponse(*rev))
}

func (h *QuestionnaireHandler) SaveDraft(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.SaveDraftRequest
	if !bindJSON(c, &req) {
		return
	}
	rev, err := h.admin.SaveDraft(c.Request.Context(), adminID, dto.ToDraftInput(req))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponse(*rev))
}

func (h *QuestionnaireHandler) DiscardDraft(c *gin.Context) {
	if err := h.admin.DiscardDraft(c.Request.Context()); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *QuestionnaireHandler) Publish(c *gin.Context) {
	rev, err := h.admin.Publish(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponse(*rev))
}

func (h *QuestionnaireHandler) ListRevisions(c *gin.Context) {
	revs, err := h.admin.ListRevisions(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponses(revs))
}

func (h *QuestionnaireHandler) GetRevision(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	rev, err := h.admin.GetRevision(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToRevisionResponse(*rev))
}

// ---- admin: nudges and review ----

func (h *QuestionnaireHandler) PreviewNudge(c *gin.Context) {
	p, err := h.admin.PreviewNudge(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToNudgePreviewResponse(*p))
}

func (h *QuestionnaireHandler) SendNudge(c *gin.Context) {
	adminID, ok := currentUser(c)
	if !ok {
		return
	}
	sent, err := h.admin.SendNudge(c.Request.Context(), adminID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.SendNudgeResponse{Sent: sent})
}

// ApplicationAnswers shows a reviewer an application's answers, private ones included, against the
// revision the applicant answered.
func (h *QuestionnaireHandler) ApplicationAnswers(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	view, err := h.questionnaire.ApplicationReviewView(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToReviewViewResponse(*view))
}
