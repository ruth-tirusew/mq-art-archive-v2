package digest

import (
	"context"
	"errors"
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
	articles   outbound.ArticleRepository
	events     outbound.EventRepository
	posts      outbound.ArtPostRepository
	recipients outbound.DigestRecipientRepository
	runs       outbound.DigestRunRepository
}

func NewService(
	articles outbound.ArticleRepository,
	events outbound.EventRepository,
	posts outbound.ArtPostRepository,
	recipients outbound.DigestRecipientRepository,
	runs outbound.DigestRunRepository,
) *Service {
	return &Service{
		articles:   articles,
		events:     events,
		posts:      posts,
		recipients: recipients,
		runs:       runs,
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
	periodEnd := now
	eventsUntil := now.Add(upcomingWindow)

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
