package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	users         outbound.UserRepository
	oauthAccounts outbound.OAuthAccountRepository
	tokens        outbound.TokenIssuer
	state         outbound.OAuthStateSigner
	providers     map[string]outbound.OAuthProvider
	allowedReturn map[string]struct{}
}

func NewService(
	users outbound.UserRepository,
	oauthAccounts outbound.OAuthAccountRepository,
	tokens outbound.TokenIssuer,
	state outbound.OAuthStateSigner,
	providers []outbound.OAuthProvider,
	allowedReturnOrigins []string,
) inbound.AuthService {
	providerMap := make(map[string]outbound.OAuthProvider, len(providers))
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	allowed := make(map[string]struct{}, len(allowedReturnOrigins))
	for _, origin := range allowedReturnOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return &Service{
		users:         users,
		oauthAccounts: oauthAccounts,
		tokens:        tokens,
		state:         state,
		providers:     providerMap,
		allowedReturn: allowed,
	}
}

func (s *Service) BeginOAuth(_ context.Context, provider, returnTo string) (string, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", "", fmt.Errorf("unknown provider: %s", provider)
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", "", err
	}

	state, err := s.state.Sign(returnTo)
	if err != nil {
		return "", "", err
	}
	return p.AuthCodeURL(state), state, nil
}

func (s *Service) CompleteOAuth(ctx context.Context, provider, code, state string) (string, *identity.User, string, error) {
	p, ok := s.providers[provider]
	if !ok {
		return "", nil, "", fmt.Errorf("unknown provider: %s", provider)
	}

	returnTo, err := s.state.Verify(state)
	if err != nil {
		return "", nil, "", err
	}
	if err := s.validateReturnTo(returnTo); err != nil {
		return "", nil, "", err
	}

	info, err := p.Exchange(ctx, code)
	if err != nil {
		return "", nil, "", err
	}
	if !info.EmailVerified {
		return "", nil, "", apperrors.ErrUnauthorized
	}

	user, err := s.findOrCreateUser(ctx, provider, info)
	if err != nil {
		return "", nil, "", err
	}

	token, err := s.tokens.Issue(ctx, user)
	if err != nil {
		return "", nil, "", err
	}
	return token, user, returnTo, nil
}

func (s *Service) findOrCreateUser(ctx context.Context, provider string, info *outbound.OAuthUserInfo) (*identity.User, error) {
	account, err := s.oauthAccounts.GetByProviderSubject(ctx, provider, info.ProviderUserID)
	if err == nil {
		return s.users.GetByID(ctx, account.UserID)
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	user, err := s.users.GetByEmail(ctx, info.Email)
	if err == nil {
		if linkErr := s.oauthAccounts.Create(ctx, identity.OAuthAccount{
			UserID:         user.ID,
			Provider:       provider,
			ProviderUserID: info.ProviderUserID,
			Email:          info.Email,
		}); linkErr != nil {
			return nil, linkErr
		}
		return user, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	newUser := identity.User{
		ID:    uuid.New(),
		Email: info.Email,
		Role:  identity.RolePublic,
	}
	if err := s.users.Create(ctx, newUser); err != nil {
		return nil, err
	}
	if err := s.oauthAccounts.Create(ctx, identity.OAuthAccount{
		UserID:         newUser.ID,
		Provider:       provider,
		ProviderUserID: info.ProviderUserID,
		Email:          info.Email,
	}); err != nil {
		return nil, err
	}
	return &newUser, nil
}

func (s *Service) validateReturnTo(returnTo string) error {
	u, err := url.Parse(returnTo)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return apperrors.ErrInvalidToken
	}
	origin := u.Scheme + "://" + u.Host
	if _, ok := s.allowedReturn[origin]; !ok {
		return apperrors.ErrForbidden
	}
	return nil
}
