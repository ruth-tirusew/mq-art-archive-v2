package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/adapter/driving/http/requestauth"
	domain "github.com/mq/api/internal/domain/settings"
	"github.com/mq/api/internal/port/inbound"
)

type SettingsHandler struct {
	settings inbound.SettingsService
}

func NewSettingsHandler(settings inbound.SettingsService) *SettingsHandler {
	return &SettingsHandler{settings: settings}
}

func (h *SettingsHandler) GetScrape(c *gin.Context) {
	view, err := h.settings.GetScrapeSettings(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toScrapeResponse(view))
}

func (h *SettingsHandler) UpdateScrape(c *gin.Context) {
	userID, err := requestauth.UserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return
	}
	var req dto.ScrapeSettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid request body"})
		return
	}
	view, err := h.settings.UpdateScrapeSettings(c.Request.Context(), userID, domain.ScrapeSettingsUpdate{
		ScrapeEnabled:         req.ScrapeEnabled,
		ScrapeSources:         req.ScrapeSources,
		ScrapeUserAgent:       req.ScrapeUserAgent,
		ScrapeTimeoutSeconds:  req.ScrapeTimeoutSeconds,
		ScrapeIntervalSeconds: req.ScrapeIntervalSeconds,
		TelegramEnabled:       req.TelegramEnabled,
		TelegramAPIID:         req.TelegramAPIID,
		TelegramAPIHash:       req.TelegramAPIHash,
		TelegramChannels:      req.TelegramChannels,
		TelegramKeywords:      req.TelegramKeywords,
		TelegramFetchLimit:    req.TelegramFetchLimit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, toScrapeResponse(view))
}

func toScrapeResponse(view *domain.ScrapeSettingsView) dto.ScrapeSettingsResponse {
	sources := view.ScrapeSources
	if sources == nil {
		sources = []string{}
	}
	channels := view.TelegramChannels
	if channels == nil {
		channels = []string{}
	}
	keywords := view.TelegramKeywords
	if keywords == nil {
		keywords = []string{}
	}
	return dto.ScrapeSettingsResponse{
		ScrapeEnabled:         view.ScrapeEnabled,
		ScrapeSources:         sources,
		ScrapeUserAgent:       view.ScrapeUserAgent,
		ScrapeTimeoutSeconds:  view.ScrapeTimeoutSeconds,
		ScrapeIntervalSeconds: view.ScrapeIntervalSeconds,
		TelegramEnabled:       view.TelegramEnabled,
		TelegramAPIID:         view.TelegramAPIID,
		TelegramAPIHashSet:    view.TelegramAPIHashSet,
		TelegramChannels:      channels,
		TelegramKeywords:      keywords,
		TelegramFetchLimit:    view.TelegramFetchLimit,
		SessionAuthorized:     view.SessionAuthorized,
	}
}
