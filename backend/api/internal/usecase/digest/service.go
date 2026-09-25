package digest

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/art"
	contentdomain "github.com/mq/api/internal/domain/content"
	digestdomain "github.com/mq/api/internal/domain/digest"
	eventsdomain "github.com/mq/api/internal/domain/events"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

// defaultLookback is the content window used when there is no prior completed run to
// measure "since" from (first-ever send).
const defaultLookback = 7 * 24 * time.Hour

// upcomingWindow bounds how far ahead events are pulled, independent of the content
// lookback: "published since last week" doesn't apply to events, which are windowed by
// when they start, not when they were created.
const upcomingWindow = 7 * 24 * time.Hour

const listLimit = 100

// Service builds digest content by composing the existing article/art/event repositories
// (same pattern as the search usecase) rather than introducing new read paths for content
// that's already listable elsewhere.
type Service struct {
	articles          outbound.ArticleRepository
	events            outbound.EventRepository
	posts             outbound.ArtPostRepository
	recipients        outbound.DigestRecipientRepository
	runs              outbound.DigestRunRepository
	notifications     outbound.NotificationPreferencesRepository
	telegramLinks     outbound.TelegramLinkRepository
	unsubscribeSecret string
	mailer            outbound.Mailer
	telegram          outbound.TelegramNotifier
	webAppURL         string
	publicAPIURL      string
}

func NewService(
	articles outbound.ArticleRepository,
	events outbound.EventRepository,
	posts outbound.ArtPostRepository,
	recipients outbound.DigestRecipientRepository,
	runs outbound.DigestRunRepository,
	notifications outbound.NotificationPreferencesRepository,
	telegramLinks outbound.TelegramLinkRepository,
	unsubscribeSecret string,
	mailer outbound.Mailer,
	telegram outbound.TelegramNotifier,
	webAppURL string,
	publicAPIURL string,
) inbound.DigestService {
	return &Service{
		articles:          articles,
		events:            events,
		posts:             posts,
		recipients:        recipients,
		runs:              runs,
		notifications:     notifications,
		telegramLinks:     telegramLinks,
		unsubscribeSecret: unsubscribeSecret,
		mailer:            mailer,
		telegram:          telegram,
		webAppURL:         strings.TrimRight(webAppURL, "/"),
		publicAPIURL:      strings.TrimRight(publicAPIURL, "/"),
	}
}

// BuildWeekly assembles the digest content window. It does not start a run or send
// anything — callers combine this with StartRun/Recipients/RecordDelivery to send.
func (s *Service) BuildWeekly(ctx context.Context) (*inbound.Digest, error) {
	now := time.Now().UTC()
	periodStart, err := s.resolvePeriodStart(ctx, now)
	if err != nil {
		return nil, err
	}
	return s.buildForPeriod(ctx, periodStart, now)
}

// buildForPeriod is BuildWeekly's content-selection logic for an explicit, already-decided
// window. SendWeekly uses this directly (rather than BuildWeekly) when resuming an
// interrupted run, so a retry rebuilds the same content window the original attempt used
// instead of a new one anchored to "now".
func (s *Service) buildForPeriod(ctx context.Context, periodStart, periodEnd time.Time) (*inbound.Digest, error) {
	eventsUntil := periodEnd.Add(upcomingWindow)

	approved := eventsdomain.EventStatusApproved
	upcomingEvents, err := s.events.List(ctx, eventsdomain.ListFilter{
		Statuses:     []eventsdomain.EventStatus{approved},
		UpcomingOnly: true,
		StartsBefore: &eventsUntil,
		Limit:        listLimit,
	})
	if err != nil {
		return nil, err
	}

	artPosts, err := s.posts.ListPublished(ctx, domain.ListFilter{
		PublishedSince: &periodStart,
		Limit:          listLimit,
	})
	if err != nil {
		return nil, err
	}

	articles, err := s.articles.ListPublished(ctx, contentdomain.ListFilter{
		PublishedSince: &periodStart,
		Limit:          listLimit,
	})
	if err != nil {
		return nil, err
	}

	return &inbound.Digest{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Events:      upcomingEvents,
		ArtPosts:    artPosts,
		Articles:    articles,
	}, nil
}

// Recipients returns everyone eligible for at least one channel. Filtering to a specific
// channel (EmailOptedIn / TelegramChatID != nil) is the sender's job, not this method's.
func (s *Service) Recipients(ctx context.Context) ([]digestdomain.Recipient, error) {
	return s.recipients.ListDigestRecipients(ctx)
}

func (s *Service) StartRun(ctx context.Context, d inbound.Digest) (*digestdomain.Run, error) {
	return s.runs.StartRun(ctx, d.PeriodStart, d.PeriodEnd)
}

func (s *Service) CompleteRun(ctx context.Context, runID uuid.UUID) error {
	return s.runs.CompleteRun(ctx, runID)
}

func (s *Service) HasSuccessfulDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel) (bool, error) {
	return s.runs.HasSuccessfulDelivery(ctx, runID, recipientID, channel)
}

func (s *Service) RecordDelivery(ctx context.Context, runID, recipientID uuid.UUID, channel digestdomain.Channel, sentAt time.Time, deliveryErr error) error {
	return s.runs.RecordDelivery(ctx, runID, recipientID, channel, sentAt, deliveryErr)
}

// SendWeekly builds the digest, sends it to every eligible recipient on every channel
// they've opted into, and records the run. It's safe to call again after a partial failure
// or a crash mid-run: if an incomplete run already exists, it's resumed under the same run
// ID (same content window, same run row) instead of starting a new one, so
// HasSuccessfulDelivery correctly reflects what was already sent and a retry only reaches
// whoever didn't get it last time. Starting a fresh run on every retry would defeat that —
// a new run ID has no delivery history, so a naive retry would resend to everyone.
//
// The Telegram message is identical for every recipient and rendered once. The email body
// is not — it carries a per-recipient signed unsubscribe link — so it's rendered fresh per
// recipient rather than reused.
func (s *Service) SendWeekly(ctx context.Context) error {
	run, err := s.currentOrNewRun(ctx)
	if err != nil {
		return fmt.Errorf("resolve digest run: %w", err)
	}

	d, err := s.buildForPeriod(ctx, run.PeriodStart, run.PeriodEnd)
	if err != nil {
		return fmt.Errorf("build weekly digest: %w", err)
	}

	recipients, err := s.Recipients(ctx)
	if err != nil {
		return fmt.Errorf("list digest recipients: %w", err)
	}

	subject := Subject(*d)
	telegramMessage := RenderTelegram(*d, s.webAppURL)

	for _, r := range recipients {
		if r.EmailOptedIn {
			s.sendChannel(ctx, run.ID, r.UserID, digestdomain.ChannelEmail, func() error {
				unsubscribeURL := fmt.Sprintf("%s/unsubscribe?token=%s", s.publicAPIURL, s.SignUnsubscribeToken(r.UserID))
				htmlBody, textBody, err := RenderEmail(*d, s.webAppURL, unsubscribeURL)
				if err != nil {
					return err
				}
				return s.mailer.SendHTML(ctx, r.Email, subject, htmlBody, textBody)
			})
		}
		if r.TelegramChatID != nil && s.telegram != nil {
			s.sendChannel(ctx, run.ID, r.UserID, digestdomain.ChannelTelegram, func() error {
				return s.telegram.Send(ctx, *r.TelegramChatID, telegramMessage)
			})
		}
	}

	return s.CompleteRun(ctx, run.ID)
}

// currentOrNewRun resumes the most recent incomplete run if one exists, otherwise starts a
// fresh one anchored to "now". An incomplete run lingering because of a genuinely stuck
// process (not just a transient crash) will keep being "resumed" indefinitely — CompleteRun
// never gets called for it — which is an acceptable failure mode here: it only delays a
// fresh period from starting, it never causes a duplicate send.
func (s *Service) currentOrNewRun(ctx context.Context) (*digestdomain.Run, error) {
	incomplete, err := s.runs.GetIncompleteRun(ctx)
	if err == nil {
		return incomplete, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	periodStart, err := s.resolvePeriodStart(ctx, now)
	if err != nil {
		return nil, err
	}
	return s.runs.StartRun(ctx, periodStart, now)
}

// sendChannel skips a recipient/channel pair already delivered in this run, otherwise
// sends and records the outcome. Send and record errors are swallowed (not returned to the
// caller) by design: one recipient's bad email address or a transient Telegram API error
// must not abort the run for everyone after them — RecordDelivery captures the failure so
// it's visible and retryable on the next run.
func (s *Service) sendChannel(ctx context.Context, runID, userID uuid.UUID, channel digestdomain.Channel, send func() error) {
	already, err := s.HasSuccessfulDelivery(ctx, runID, userID, channel)
	if err != nil || already {
		return
	}
	sendErr := send()
	_ = s.RecordDelivery(ctx, runID, userID, channel, time.Now().UTC(), sendErr)
}

// resolvePeriodStart anchors the content window to the end of the last completed run, so
// consecutive sends cover contiguous, non-overlapping periods. Falls back to a fixed
// lookback when there's no prior run (first send) or the lookup fails for any other reason
// than "none exists" — a transient DB error here shouldn't silently produce an empty digest.
func (s *Service) resolvePeriodStart(ctx context.Context, now time.Time) (time.Time, error) {
	last, err := s.runs.LastCompletedRun(ctx)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return now.Add(-defaultLookback), nil
		}
		return time.Time{}, err
	}
	return last.PeriodEnd, nil
}

// linkTokenTTL is short — this token only has to survive the few seconds between the
// settings page generating a t.me deep link and the user tapping it in their Telegram app.
const linkTokenTTL = 15 * time.Minute

// CreateTelegramLink issues a one-time token for the /start deep link
// (t.me/<bot>?start=<token>) that links userID's account to whichever chat sends it.
func (s *Service) CreateTelegramLink(ctx context.Context, userID uuid.UUID) (string, error) {
	token, err := randomToken(24)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	if err := s.telegramLinks.Create(ctx, digestdomain.LinkToken{
		Token:     token,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(linkTokenTTL),
	}); err != nil {
		return "", err
	}
	return token, nil
}

// HandleTelegramStart handles a "/start <token>" message from the bot's update loop,
// linking the sending chat to whichever account the token belongs to.
func (s *Service) HandleTelegramStart(ctx context.Context, chatID, text string) error {
	token := parseStartToken(text)
	if token == "" {
		return fmt.Errorf("%w: missing start token", apperrors.ErrValidation)
	}
	link, err := s.telegramLinks.Consume(ctx, token)
	if err != nil {
		return err
	}
	return s.notifications.SetTelegramChatID(ctx, link.UserID, &chatID)
}

// HandleTelegramStop handles a "/stop" message, unlinking the sending chat. A chat with no
// linked account is treated as already-stopped rather than an error.
func (s *Service) HandleTelegramStop(ctx context.Context, chatID string) error {
	prefs, err := s.notifications.GetByTelegramChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	}
	return s.notifications.SetTelegramChatID(ctx, prefs.UserID, nil)
}

func parseStartToken(text string) string {
	const prefix = "/start"
	if !strings.HasPrefix(text, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(text, prefix))
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// SignUnsubscribeToken produces a stateless, non-expiring token for an email's unsubscribe
// link. Unlike the Telegram link token or the password-reset/email-verification tokens
// elsewhere in this codebase, it deliberately isn't stored or short-lived: a digest email
// can sit unread for weeks, and its unsubscribe link must still work whenever it's opened.
func (s *Service) SignUnsubscribeToken(userID uuid.UUID) string {
	return userID.String() + "." + base64.RawURLEncoding.EncodeToString(s.unsubscribeSignature(userID.String()))
}

// Unsubscribe verifies a token from SignUnsubscribeToken and turns off the newsletter flag
// for that user. Verification failures and an already-unsubscribed user both return nil —
// this endpoint is meant to always show "you're unsubscribed", not leak which tokens are
// valid.
func (s *Service) Unsubscribe(ctx context.Context, token string) error {
	userID, err := s.verifyUnsubscribeToken(token)
	if err != nil {
		return nil
	}
	prefs, err := s.notifications.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		return err
	}
	prefs.NewsletterEnabled = false
	return s.notifications.Upsert(ctx, *prefs)
}

func (s *Service) verifyUnsubscribeToken(token string) (uuid.UUID, error) {
	rawID, encodedSig, found := strings.Cut(token, ".")
	if !found {
		return uuid.Nil, apperrors.ErrValidation
	}
	userID, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, apperrors.ErrValidation
	}
	sig, err := base64.RawURLEncoding.DecodeString(encodedSig)
	if err != nil {
		return uuid.Nil, apperrors.ErrValidation
	}
	if !hmac.Equal(sig, s.unsubscribeSignature(rawID)) {
		return uuid.Nil, apperrors.ErrValidation
	}
	return userID, nil
}

func (s *Service) unsubscribeSignature(rawUserID string) []byte {
	mac := hmac.New(sha256.New, []byte(s.unsubscribeSecret))
	mac.Write([]byte(rawUserID))
	return mac.Sum(nil)
}
