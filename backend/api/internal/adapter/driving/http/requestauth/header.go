package requestauth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
)

const (
	ContextUserID    = "user_id"
	ContextUserRole  = "user_role"
	ContextUserEmail = "user_email"
)

var (
	ErrUnauthorized = apperrors.ErrUnauthorized
	ErrInvalidToken = apperrors.ErrInvalidToken
)

// UserIDFromHeader extracts a stub user ID from X-User-ID for local development.
func UserIDFromHeader(header string) (uuid.UUID, error) {
	if header == "" {
		return uuid.Nil, ErrUnauthorized
	}
	id, err := uuid.Parse(header)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return id, nil
}

func UserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	raw, ok := c.Get(ContextUserID)
	if !ok {
		return uuid.Nil, ErrUnauthorized
	}
	id, ok := raw.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, ErrUnauthorized
	}
	return id, nil
}

func UserRoleFromContext(c *gin.Context) (identity.Role, error) {
	raw, ok := c.Get(ContextUserRole)
	if !ok {
		return "", ErrUnauthorized
	}
	role, ok := raw.(identity.Role)
	if !ok || role == "" {
		return "", ErrUnauthorized
	}
	return role, nil
}

func UnauthorizedError() error {
	return errors.Join(ErrUnauthorized)
}
