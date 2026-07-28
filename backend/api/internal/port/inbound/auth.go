package inbound

import (
	"context"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/identity"
)

type AuthService interface {
	BeginOAuth(ctx context.Context, provider, returnTo string) (authURL string, state string, err error)
	CompleteOAuth(ctx context.Context, provider, code, state string) (accessToken string, user *identity.User, returnTo string, err error)
	Register(ctx context.Context, email, password string) (accessToken string, user *identity.User, err error)
	Login(ctx context.Context, email, password string) (accessToken string, user *identity.User, err error)
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	VerifyEmail(ctx context.Context, token string) error
	ResendEmailVerification(ctx context.Context, userID uuid.UUID) error

	GetMe(ctx context.Context, userID uuid.UUID) (*identity.User, error)
	UpdateMyProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL string) (*identity.User, error)
	ChangeEmail(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) (*identity.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error

	GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs identity.NotificationPreferences) (*identity.NotificationPreferences, error)
	ListUsers(ctx context.Context, role *identity.Role, limit, offset int) ([]identity.User, int, error)
	UpdateUserRole(ctx context.Context, actorID, userID uuid.UUID, role identity.Role) (*identity.User, error)
}
