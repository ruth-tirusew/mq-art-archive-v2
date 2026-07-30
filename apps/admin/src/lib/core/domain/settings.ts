export interface NotificationPreferences {
  email_on_new_application: boolean;
  email_on_event_sync_summary: boolean;
  newsletter_enabled: boolean;
}

export interface ScrapeSettings {
  scrape_enabled: boolean;
  scrape_sources: string[];
  scrape_user_agent: string;
  scrape_timeout_seconds: number;
  scrape_interval_seconds: number;
  telegram_enabled: boolean;
  telegram_api_id: number;
  telegram_api_hash_set: boolean;
  telegram_channels: string[];
  telegram_keywords: string[];
  telegram_fetch_limit: number;
  session_authorized: boolean;
}

export interface ScrapeSettingsUpdate {
  scrape_enabled?: boolean;
  scrape_sources?: string[];
  scrape_user_agent?: string;
  scrape_timeout_seconds?: number;
  scrape_interval_seconds?: number;
  telegram_enabled?: boolean;
  telegram_api_id?: number;
  telegram_api_hash?: string;
  telegram_channels?: string[];
  telegram_keywords?: string[];
  telegram_fetch_limit?: number;
}

export interface UpdateProfileInput {
  display_name: string;
  avatar_url: string;
}
