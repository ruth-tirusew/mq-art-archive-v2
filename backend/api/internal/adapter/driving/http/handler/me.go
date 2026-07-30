package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	domain "github.com/mq/api/internal/domain/profile"
)

func (h *ProfileHandler) GetMyProfile(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	profile, err := h.profiles.GetArtistByUserID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*profile))
}

func (h *ProfileHandler) UpdateMyProfile(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req dto.UpdateArtistProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	profile, err := h.profiles.UpdateOwnProfile(c.Request.Context(), userID, req.ToOwnProfileUpdate())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtistProfileResponse(*profile))
}

func (h *ArtHandler) ListMyPosts(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}

	posts, err := h.art.ListOwnedByArtist(c.Request.Context(), artist.ID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponses(posts))
}

func (h *ArtHandler) GetMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	post, err := h.art.GetOwned(c.Request.Context(), artist.ID, postID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) UpdateMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	var req dto.UpdateArtPostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	post, err := h.art.UpdateOwned(c.Request.Context(), artist.ID, postID, req.ToWrite())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) PublishMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	post, err := h.art.PublishOwned(c.Request.Context(), artist.ID, postID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) UnpublishMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	post, err := h.art.UnpublishOwned(c.Request.Context(), artist.ID, postID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) ArchiveMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	post, err := h.art.ArchiveOwned(c.Request.Context(), artist.ID, postID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToArtPostResponse(*post))
}

func (h *ArtHandler) DeleteMyPost(c *gin.Context) {
	artist, ok := h.requireArtist(c)
	if !ok {
		return
	}
	postID, err := uuidParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return
	}
	if err := h.art.DeleteOwned(c.Request.Context(), artist.ID, postID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ArtHandler) requireArtist(c *gin.Context) (*domain.ArtistProfile, bool) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return nil, false
	}
	artist, err := h.profile.GetArtistByUserID(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return nil, false
	}
	return artist, true
}
