package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/outbound"
)

type accessClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret    []byte
	accessTTL time.Duration
}

func NewTokenService(secret string, accessTTL time.Duration) *TokenService {
	return &TokenService{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
}

func (s *TokenService) Issue(_ context.Context, user *identity.User) (string, error) {
	now := time.Now().UTC()
	claims := accessClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (s *TokenService) Verify(_ context.Context, tokenString string) (outbound.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, apperrors.ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return outbound.TokenClaims{}, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*accessClaims)
	if !ok || claims.Subject == "" {
		return outbound.TokenClaims{}, apperrors.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return outbound.TokenClaims{}, apperrors.ErrInvalidToken
	}

	return outbound.TokenClaims{
		UserID: userID,
		Role:   identity.Role(claims.Role),
	}, nil
}

// OAuthStateService signs short-lived state tokens for CSRF protection.
type OAuthStateService struct {
	secret []byte
	ttl    time.Duration
}

const statePurposeLink = "link"

type oauthStateClaims struct {
	ReturnTo string `json:"return_to"`
	// Purpose is empty for a normal sign-in and "link" when a signed-in user is connecting a
	// provider to their own account. UserID binds a link state to that user.
	Purpose string `json:"purpose,omitempty"`
	UserID  string `json:"uid,omitempty"`
	jwt.RegisteredClaims
}

func NewOAuthStateService(secret string, ttl time.Duration) *OAuthStateService {
	return &OAuthStateService{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (s *OAuthStateService) Sign(returnTo string) (string, error) {
	now := time.Now().UTC()
	claims := oauthStateClaims{
		ReturnTo: returnTo,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *OAuthStateService) parse(state string) (*oauthStateClaims, error) {
	token, err := jwt.ParseWithClaims(state, &oauthStateClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, apperrors.ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}
	claims, ok := token.Claims.(*oauthStateClaims)
	if !ok || claims.ReturnTo == "" {
		return nil, apperrors.ErrInvalidToken
	}
	return claims, nil
}

// Verify accepts only sign-in states. A link state must not be able to start a session.
func (s *OAuthStateService) Verify(state string) (string, error) {
	claims, err := s.parse(state)
	if err != nil || claims.Purpose != "" {
		return "", apperrors.ErrInvalidToken
	}
	return claims.ReturnTo, nil
}

// SignLink creates a state that is only valid for linking a provider to userID.
func (s *OAuthStateService) SignLink(returnTo string, userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := oauthStateClaims{
		ReturnTo: returnTo,
		Purpose:  statePurposeLink,
		UserID:   userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// VerifyLink accepts only link states and returns the user they were issued for.
func (s *OAuthStateService) VerifyLink(state string) (string, uuid.UUID, error) {
	claims, err := s.parse(state)
	if err != nil || claims.Purpose != statePurposeLink {
		return "", uuid.Nil, apperrors.ErrInvalidToken
	}
	id, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", uuid.Nil, apperrors.ErrInvalidToken
	}
	return claims.ReturnTo, id, nil
}
