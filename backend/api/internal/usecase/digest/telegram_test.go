package digest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	digestdomain "github.com/mq/api/internal/domain/digest"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/testutil/assist"
)

type fakeNotificationsRepo struct {
	byUserID    map[uuid.UUID]identity.NotificationPreferences
	byChatID    map[string]uuid.UUID
	upsertCalls []identity.NotificationPreferences
}

func newFakeNotificationsRepo() *fakeNotificationsRepo {
	return &fakeNotificationsRepo{
		byUserID: make(map[uuid.UUID]identity.NotificationPreferences),
		byChatID: make(map[string]uuid.UUID),
	}
}

func (f *fakeNotificationsRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*identity.NotificationPreferences, error) {
	prefs, ok := f.byUserID[userID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return &prefs, nil
}

func (f *fakeNotificationsRepo) Upsert(ctx context.Context, prefs identity.NotificationPreferences) error {
	f.upsertCalls = append(f.upsertCalls, prefs)
	f.byUserID[prefs.UserID] = prefs
	if prefs.TelegramChatID != nil {
		f.byChatID[*prefs.TelegramChatID] = prefs.UserID
	}
	return nil
}

func (f *fakeNotificationsRepo) SetTelegramChatID(ctx context.Context, userID uuid.UUID, chatID *string) error {
	prefs, ok := f.byUserID[userID]
	if !ok {
		prefs = identity.DefaultNotificationPreferences(userID)
	}
	if prefs.TelegramChatID != nil {
		delete(f.byChatID, *prefs.TelegramChatID)
	}
	prefs.TelegramChatID = chatID
	f.byUserID[userID] = prefs
	if chatID != nil {
		f.byChatID[*chatID] = userID
	}
	return nil
}

func (f *fakeNotificationsRepo) GetByTelegramChatID(ctx context.Context, chatID string) (*identity.NotificationPreferences, error) {
	userID, ok := f.byChatID[chatID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	prefs := f.byUserID[userID]
	return &prefs, nil
}

type fakeTelegramLinkRepo struct {
	tokens map[string]digestdomain.LinkToken
}

func newFakeTelegramLinkRepo() *fakeTelegramLinkRepo {
	return &fakeTelegramLinkRepo{tokens: make(map[string]digestdomain.LinkToken)}
}

func (f *fakeTelegramLinkRepo) Create(ctx context.Context, token digestdomain.LinkToken) error {
	f.tokens[token.Token] = token
	return nil
}

func (f *fakeTelegramLinkRepo) Consume(ctx context.Context, token string) (*digestdomain.LinkToken, error) {
	lt, ok := f.tokens[token]
	if !ok || lt.ConsumedAt != nil || time.Now().After(lt.ExpiresAt) {
		return nil, apperrors.ErrNotFound
	}
	now := time.Now().UTC()
	lt.ConsumedAt = &now
	f.tokens[token] = lt
	return &lt, nil
}

func newTestService(notifications *fakeNotificationsRepo, links *fakeTelegramLinkRepo) *Service {
	return &Service{notifications: notifications, telegramLinks: links, unsubscribeSecret: "test-secret"}
}

func TestCreateTelegramLink_thenHandleStart_linksChat(t *testing.T) {
	notifications := newFakeNotificationsRepo()
	links := newFakeTelegramLinkRepo()
	svc := newTestService(notifications, links)
	userID := uuid.New()
	ctx := context.Background()

	token, err := svc.CreateTelegramLink(ctx, userID)
	assist.NoError(t, err)

	assist.NoError(t, svc.HandleTelegramStart(ctx, "12345", "/start "+token))

	prefs, err := notifications.GetByTelegramChatID(ctx, "12345")
	assist.NoError(t, err)
	assist.Equal(t, userID, prefs.UserID)
}

func TestHandleTelegramStart_rejectsMissingToken(t *testing.T) {
	svc := newTestService(newFakeNotificationsRepo(), newFakeTelegramLinkRepo())
	err := svc.HandleTelegramStart(context.Background(), "12345", "/start")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestHandleTelegramStart_rejectsAlreadyConsumedToken(t *testing.T) {
	notifications := newFakeNotificationsRepo()
	links := newFakeTelegramLinkRepo()
	svc := newTestService(notifications, links)
	userID := uuid.New()
	ctx := context.Background()

	token, err := svc.CreateTelegramLink(ctx, userID)
	assist.NoError(t, err)
	assist.NoError(t, svc.HandleTelegramStart(ctx, "12345", "/start "+token))

	// A second chat replaying the same token must not also link.
	err = svc.HandleTelegramStart(ctx, "99999", "/start "+token)
	assist.ErrorIs(t, err, apperrors.ErrNotFound)

	_, err = notifications.GetByTelegramChatID(ctx, "99999")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestHandleTelegramStop_unlinksChat(t *testing.T) {
	notifications := newFakeNotificationsRepo()
	svc := newTestService(notifications, newFakeTelegramLinkRepo())
	userID := uuid.New()
	ctx := context.Background()

	assist.NoError(t, notifications.SetTelegramChatID(ctx, userID, ptr("12345")))
	assist.NoError(t, svc.HandleTelegramStop(ctx, "12345"))

	_, err := notifications.GetByTelegramChatID(ctx, "12345")
	assist.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestHandleTelegramStop_unknownChatIsNotAnError(t *testing.T) {
	svc := newTestService(newFakeNotificationsRepo(), newFakeTelegramLinkRepo())
	assist.NoError(t, svc.HandleTelegramStop(context.Background(), "does-not-exist"))
}

func TestSignAndVerifyUnsubscribeToken_disablesNewsletter(t *testing.T) {
	notifications := newFakeNotificationsRepo()
	svc := newTestService(notifications, newFakeTelegramLinkRepo())
	userID := uuid.New()
	ctx := context.Background()

	prefs := identity.DefaultNotificationPreferences(userID)
	prefs.NewsletterEnabled = true
	assist.NoError(t, notifications.Upsert(ctx, prefs))

	token := svc.SignUnsubscribeToken(userID)
	assist.NoError(t, svc.Unsubscribe(ctx, token))

	got, err := notifications.GetByUserID(ctx, userID)
	assist.NoError(t, err)
	assist.Equal(t, false, got.NewsletterEnabled)
}

func TestUnsubscribe_tamperedTokenChangesNothing(t *testing.T) {
	notifications := newFakeNotificationsRepo()
	svc := newTestService(notifications, newFakeTelegramLinkRepo())
	userID := uuid.New()
	ctx := context.Background()

	prefs := identity.DefaultNotificationPreferences(userID)
	prefs.NewsletterEnabled = true
	assist.NoError(t, notifications.Upsert(ctx, prefs))

	forgedToken := userID.String() + ".not-a-real-signature"
	assist.NoError(t, svc.Unsubscribe(ctx, forgedToken))

	got, err := notifications.GetByUserID(ctx, userID)
	assist.NoError(t, err)
	assist.Equal(t, true, got.NewsletterEnabled)
}

func ptr(s string) *string { return &s }
