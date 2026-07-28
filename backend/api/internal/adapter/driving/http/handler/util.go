package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driving/http/dto"
	"github.com/mq/api/internal/domain/apperrors"
)

func uuidParam(c *gin.Context, name string) (uuid.UUID, error) {
	return uuid.Parse(c.Param(name))
}

func writeError(c *gin.Context, err error) {
	method, path := "", ""
	if c.Request != nil {
		method, path = c.Request.Method, c.Request.URL.Path
	}
	log.Printf("request_error request_id=%q method=%q path=%q error=%q", c.GetString("request_id"), method, path, err.Error())
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrNotImplemented):
		c.JSON(http.StatusNotImplemented, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, apperrors.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
	}
}
