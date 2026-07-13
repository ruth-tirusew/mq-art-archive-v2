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

type oauthStateClaims struct {
	ReturnTo string `json:"return_to"`
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

func (s *OAuthStateService) Verify(state string) (string, error) {
	token, err := jwt.ParseWithClaims(state, &oauthStateClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, apperrors.ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*oauthStateClaims)
	if !ok || claims.ReturnTo == "" {
		return "", apperrors.ErrInvalidToken
	}
	return claims.ReturnTo, nil
}
