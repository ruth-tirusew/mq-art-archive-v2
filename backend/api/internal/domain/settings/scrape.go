package settings

import (
	"time"

	"github.com/google/uuid"
)

// ScrapeSettings is the singleton operational scrape/MTProto configuration.
type ScrapeSettings struct {
	ScrapeEnabled         bool
	ScrapeSources         []string
	ScrapeUserAgent       string
	ScrapeTimeoutSeconds  int
	ScrapeIntervalSeconds int

	TelegramEnabled    bool
	TelegramAPIID      int
	TelegramAPIHash    string
	TelegramChannels   []string
	TelegramKeywords   []string
	TelegramFetchLimit int

	UpdatedAt time.Time
	UpdatedBy *uuid.UUID
}

// ScrapeSettingsView is the admin-safe view (secrets masked).
type ScrapeSettingsView struct {
	ScrapeEnabled         bool
	ScrapeSources         []string
	ScrapeUserAgent       string
	ScrapeTimeoutSeconds  int
	ScrapeIntervalSeconds int

	TelegramEnabled    bool
	TelegramAPIID      int
	TelegramAPIHashSet bool
	TelegramChannels   []string
	TelegramKeywords   []string
	TelegramFetchLimit int
	SessionAuthorized  bool

	UpdatedAt time.Time
}

// ScrapeSettingsUpdate is a partial update from admin UI.
type ScrapeSettingsUpdate struct {
	ScrapeEnabled         *bool
	ScrapeSources         *[]string
	ScrapeUserAgent       *string
	ScrapeTimeoutSeconds  *int
	ScrapeIntervalSeconds *int

	TelegramEnabled    *bool
	TelegramAPIID      *int
	TelegramAPIHash    *string // nil = leave unchanged; empty string clears
	TelegramChannels   *[]string
	TelegramKeywords   *[]string
	TelegramFetchLimit *int
}
