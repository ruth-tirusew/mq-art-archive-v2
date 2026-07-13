package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
)

func TestTokenService_IssueAndVerify(t *testing.T) {
	svc := NewTokenService("test-secret", time.Hour)
	userID := uuid.New()
	user := &identity.User{
		ID:    userID,
		Email: "user@example.com",
		Role:  identity.RoleArtist,
	}

	token, err := svc.Issue(t.Context(), user)
	assist.NoError(t, err)

	claims, err := svc.Verify(t.Context(), token)
	assist.NoError(t, err)
	assist.Equal(t, userID, claims.UserID)
	assist.Equal(t, identity.RoleArtist, claims.Role)
}

func TestTokenService_Verify_invalidToken(t *testing.T) {
	svc := NewTokenService("test-secret", time.Hour)
	_, err := svc.Verify(t.Context(), "not-a-jwt")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrInvalidToken)
}

func TestOAuthStateService_SignAndVerify(t *testing.T) {
	svc := NewOAuthStateService("state-secret", 10*time.Minute)
	state, err := svc.Sign("http://localhost:5173/auth/callback")
	assist.NoError(t, err)

	returnTo, err := svc.Verify(state)
	assist.NoError(t, err)
	assist.Equal(t, "http://localhost:5173/auth/callback", returnTo)
}

func TestOAuthStateService_Verify_tampered(t *testing.T) {
	svc := NewOAuthStateService("state-secret", 10*time.Minute)
	_, err := svc.Verify("tampered")
	assist.Error(t, err)
	assist.ErrorIs(t, err, apperrors.ErrInvalidToken)
}
