// Package digest holds the digest-specific domain concepts that don't cross into other
// bounded contexts. The content aggregate (which does span art/content/events) lives in
// internal/port/inbound instead, alongside SearchResults, since domain packages here may
// not import one another — see internal/arch's TestArchitecture.
package digest

import (
	"time"

	"github.com/google/uuid"
)

// Channel identifies a delivery channel for a digest send.
type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelTelegram Channel = "telegram"
)

// Recipient is a subscriber eligible for at least one channel. EmailOptedIn distinguishes
// "has an email" (every user does) from "wants the digest by email" (NewsletterEnabled).
type Recipient struct {
	UserID         uuid.UUID
	Email          string
	EmailOptedIn   bool
	TelegramChatID *string
}

// Run is one digest send attempt, used to compute the next period and to make sends
// idempotent: a retried run can skip recipients already delivered to (see Delivery).
type Run struct {
	ID          uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
	StartedAt   time.Time
	CompletedAt *time.Time
}

// LinkToken is a one-time token used to link a Telegram chat to a site account: the
// settings page generates one and shows a t.me/<bot>?start=<token> link, and the bot
// resolves it back to the user when it receives that /start command.
type LinkToken struct {
	Token      string
	UserID     uuid.UUID
	CreatedAt  time.Time
	ExpiresAt  time.Time
	ConsumedAt *time.Time
}
