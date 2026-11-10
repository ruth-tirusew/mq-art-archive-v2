package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/port/inbound"
)

// ApplicantWorksHandler is the work an applicant uploads to show with their application. Any
// signed-in user can use it: applicants are not artists yet.
type ApplicantWorksHandler struct{ works inbound.ApplicantWorksService }

func NewApplicantWorksHandler(works inbound.ApplicantWorksService) *ApplicantWorksHandler {
	return &ApplicantWorksHandler{works: works}
}

func (h *ApplicantWorksHandler) List(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	works, err := h.works.ListWorks(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	out := dto.ToApplicantWorkResponses(works)
	if out == nil {
		out = []dto.ApplicantWorkResponse{}
	}
	c.JSON(http.StatusOK, out)
}

func (h *ApplicantWorksHandler) Add(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.AddWorkRequest
	if !bindJSON(c, &req) {
		return
	}
	work, err := h.works.AddWork(c.Request.Context(), userID, inbound.AddWorkInput{
		MediaAssetID: req.MediaAssetID, Title: req.Title, Medium: req.Medium, Year: req.Year,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.ToApplicantWorkResponse(*work))
}

func (h *ApplicantWorksHandler) Remove(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.works.RemoveWork(c.Request.Context(), userID, id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ApplicantWorksHandler) GetPortrait(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	portrait, err := h.works.GetPortrait(c.Request.Context(), userID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToApplicantPortraitResponse(portrait))
}

func (h *ApplicantWorksHandler) SetPortrait(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.SetPortraitRequest
	if !bindJSON(c, &req) {
		return
	}
	portrait, err := h.works.SetPortrait(c.Request.Context(), userID, req.MediaAssetID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToApplicantPortraitResponse(portrait))
}

func (h *ApplicantWorksHandler) RemovePortrait(c *gin.Context) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	if err := h.works.RemovePortrait(c.Request.Context(), userID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
