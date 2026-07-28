package dto

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
	DisplayName   string `json:"display_name,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	HasPassword   bool   `json:"has_password"`
}

type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type ChangeEmailRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type NotificationPreferencesResponse struct {
	EmailOnNewApplication   bool `json:"email_on_new_application"`
	EmailOnEventSyncSummary bool `json:"email_on_event_sync_summary"`
	NewsletterEnabled       bool `json:"newsletter_enabled"`
}

type ScrapeSettingsResponse struct {
	ScrapeEnabled         bool     `json:"scrape_enabled"`
	ScrapeSources         []string `json:"scrape_sources"`
	ScrapeUserAgent       string   `json:"scrape_user_agent"`
	ScrapeTimeoutSeconds  int      `json:"scrape_timeout_seconds"`
	ScrapeIntervalSeconds int      `json:"scrape_interval_seconds"`

	TelegramEnabled    bool     `json:"telegram_enabled"`
	TelegramAPIID      int      `json:"telegram_api_id"`
	TelegramAPIHashSet bool     `json:"telegram_api_hash_set"`
	TelegramChannels   []string `json:"telegram_channels"`
	TelegramKeywords   []string `json:"telegram_keywords"`
	TelegramFetchLimit int      `json:"telegram_fetch_limit"`
	SessionAuthorized  bool     `json:"session_authorized"`
}

type ScrapeSettingsUpdateRequest struct {
	ScrapeEnabled         *bool     `json:"scrape_enabled"`
	ScrapeSources         *[]string `json:"scrape_sources"`
	ScrapeUserAgent       *string   `json:"scrape_user_agent"`
	ScrapeTimeoutSeconds  *int      `json:"scrape_timeout_seconds"`
	ScrapeIntervalSeconds *int      `json:"scrape_interval_seconds"`

	TelegramEnabled    *bool     `json:"telegram_enabled"`
	TelegramAPIID      *int      `json:"telegram_api_id"`
	TelegramAPIHash    *string   `json:"telegram_api_hash"`
	TelegramChannels   *[]string `json:"telegram_channels"`
	TelegramKeywords   *[]string `json:"telegram_keywords"`
	TelegramFetchLimit *int      `json:"telegram_fetch_limit"`
}
