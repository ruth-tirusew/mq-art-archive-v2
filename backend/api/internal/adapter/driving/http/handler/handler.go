package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}

type ArticleHandler struct {
	content inbound.ContentService
}

func NewArticleHandler(content inbound.ContentService) *ArticleHandler {
	return &ArticleHandler{content: content}
}

func (h *ArticleHandler) List(c *gin.Context) {
	articles, err := h.content.ListPublished(c.Request.Context(), articleListFilter(c))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponses(articles))
}

func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	article, err := h.content.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponse(*article))
}

func (h *ArticleHandler) Create(c *gin.Context) {
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	authorID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	article, err := h.content.CreateDraft(c.Request.Context(), authorID, req.Title, req.Body)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToArticleResponse(*article))
}

type ProfileHandler struct {
	profiles inbound.ProfileService
}

func NewProfileHandler(profiles inbound.ProfileService) *ProfileHandler {
	return &ProfileHandler{profiles: profiles}
}

func (h *ProfileHandler) GetBySlug(c *gin.Context) {
	profile, err := h.profiles.GetArtistBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*profile))
}

type InstitutionHandler struct {
	institutions inbound.InstitutionService
}

func NewInstitutionHandler(institutions inbound.InstitutionService) *InstitutionHandler {
	return &InstitutionHandler{institutions: institutions}
}

func (h *InstitutionHandler) GetBySlug(c *gin.Context) {
	inst, err := h.institutions.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, inst)
}

type ArtHandler struct {
	art     inbound.ArtService
	profile inbound.ProfileService
}

func NewArtHandler(art inbound.ArtService, profile inbound.ProfileService) *ArtHandler {
	return &ArtHandler{art: art, profile: profile}
}

func (h *ArtHandler) ListByArtistSlug(c *gin.Context) {
	artist, err := h.profile.GetArtistBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	posts, err := h.art.ListByArtist(c.Request.Context(), artist.ID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponses(posts))
}

func (h *ArtHandler) GetByID(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	post, err := h.art.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) Create(c *gin.Context) {
	var req dto.CreateArtPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	artist, err := h.profile.GetArtistByUserID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}

	post, err := h.art.CreateDraft(c.Request.Context(), artist.ID, req.Title, req.Description, req.Medium)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToArtPostResponse(*post))
}

type OnboardingHandler struct {
	onboarding inbound.OnboardingService
}

func NewOnboardingHandler(onboarding inbound.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{onboarding: onboarding}
}

func (h *OnboardingHandler) ListPending(c *gin.Context) {
	apps, err := h.onboarding.ListPending(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, apps)
}

func (h *OnboardingHandler) GetByID(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	app, err := h.onboarding.GetByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *OnboardingHandler) Review(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}

	var req dto.ReviewApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	status := onboarding.ApprovalStatus(req.Status)
	if status != onboarding.ApprovalStatusApproved && status != onboarding.ApprovalStatusRejected {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "status must be approved or rejected"})
		return
	}

	reviewerID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	app, err := h.onboarding.Review(c.Request.Context(), id, reviewerID, status, req.Notes)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, app)
}
