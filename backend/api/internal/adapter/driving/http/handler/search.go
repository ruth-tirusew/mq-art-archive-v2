package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/port/inbound"
)

type SearchHandler struct {
	search inbound.SearchService
}

func NewSearchHandler(search inbound.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "q is required"})
		return
	}

	results, err := h.search.Search(c.Request.Context(), q, queryLimit(c, 20))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.ToSearchResponse(results.Articles, results.Events, results.Artists, results.Posts))
}
