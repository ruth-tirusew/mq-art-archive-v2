package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/content"
	"github.com/mq/api/internal/domain/profile"
)

type AdminUpdateArtistRequest struct {
	Status   *string `json:"status"`
	Featured *bool   `json:"featured"`
}

func (h *ProfileHandler) ListArtistsAdmin(c *gin.Context) {
	var status *profile.ProfileStatus
	if raw := c.Query("status"); raw != "" {
		s := profile.ProfileStatus(raw)
		status = &s
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	profiles, err := h.profiles.ListAll(c.Request.Context(), status, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponses(profiles))
}

func (h *ProfileHandler) GetArtistAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	p, err := h.profiles.GetArtistByID(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*p))
}

func (h *ProfileHandler) CreateArtistAdmin(c *gin.Context) {
	var req dto.AdminArtistWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	p, err := h.profiles.AdminCreate(c.Request.Context(), req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToArtistProfileResponse(*p))
}

func (h *ProfileHandler) UpdateArtistAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.AdminArtistWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	p, err := h.profiles.AdminUpdateContent(c.Request.Context(), id, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*p))
}

func (h *ProfileHandler) DeleteArtistAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := h.profiles.AdminDelete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProfileHandler) PatchArtistAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req AdminUpdateArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	var status *profile.ProfileStatus
	if req.Status != nil {
		s := profile.ProfileStatus(*req.Status)
		status = &s
	}
	p, err := h.profiles.AdminUpdate(c.Request.Context(), id, status, req.Featured)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*p))
}

func (h *ArtHandler) ListPostsAdmin(c *gin.Context) {
	var status *art.ArtStatus
	if raw := c.Query("status"); raw != "" {
		s := art.ArtStatus(raw)
		status = &s
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, err := h.art.ListAll(c.Request.Context(), status, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostWithArtistResponses(posts))
}

func (h *ArtHandler) GetPostAdmin(c *gin.Context) {
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

func (h *ArtHandler) CreatePostAdmin(c *gin.Context) {
	var req dto.AdminCreateArtPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	var status *art.ArtStatus
	if req.Status != nil && *req.Status != "" {
		s := art.ArtStatus(*req.Status)
		status = &s
	}
	post, err := h.art.AdminCreate(c.Request.Context(), req.ArtistID, req.ToWrite(), status)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) UpdatePostAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.UpdateArtPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	post, err := h.art.AdminUpdateContent(c.Request.Context(), id, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) DeletePostAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := h.art.AdminDelete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ArtHandler) PatchPostAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.AdminUpdateArtPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	var status *art.ArtStatus
	if req.Status != nil {
		s := art.ArtStatus(*req.Status)
		status = &s
	}
	post, err := h.art.AdminUpdate(c.Request.Context(), id, status, req.FeaturedAcquisition)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArticleHandler) ListArticlesAdmin(c *gin.Context) {
	var status *content.ArticleStatus
	if raw := c.Query("status"); raw != "" {
		s := content.ArticleStatus(raw)
		status = &s
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	articles, err := h.content.AdminList(c.Request.Context(), status, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponses(articles))
}

func (h *ArticleHandler) GetArticleAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	article, err := h.content.AdminGet(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponse(*article))
}

func (h *ArticleHandler) CreateArticleAdmin(c *gin.Context) {
	var req dto.AdminArticleWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	authorID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	article, err := h.content.AdminCreate(c.Request.Context(), authorID, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToArticleResponse(*article))
}

func (h *ArticleHandler) UpdateArticleAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.AdminArticleWriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	editorID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	article, err := h.content.AdminUpdate(c.Request.Context(), id, editorID, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponse(*article))
}

func (h *ArticleHandler) ListArticleRevisionsAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	revs, err := h.content.AdminListRevisions(c.Request.Context(), id, limit, offset)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleRevisionResponses(revs))
}

func (h *ArticleHandler) GetArticleRevisionAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil || version < 1 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid version"})
		return
	}
	rev, err := h.content.AdminGetRevision(c.Request.Context(), id, version)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleRevisionResponse(*rev))
}

func (h *ArticleHandler) RestoreArticleRevisionAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	version, err := strconv.Atoi(c.Param("version"))
	if err != nil || version < 1 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid version"})
		return
	}
	editorID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	article, err := h.content.AdminRestoreRevision(c.Request.Context(), id, version, editorID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponse(*article))
}

func (h *ArticleHandler) PatchArticleAdmin(c *gin.Context) {
	id, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.AdminPatchArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	var status *content.ArticleStatus
	if req.Status != nil {
		s := content.ArticleStatus(*req.Status)
		status = &s
	}
	article, err := h.content.AdminSetStatus(c.Request.Context(), id, status, req.Verified)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArticleResponse(*article))
}
