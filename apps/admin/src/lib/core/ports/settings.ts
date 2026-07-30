import type { ScrapeSettings, ScrapeSettingsUpdate } from '$lib/core/domain/settings';

export interface SettingsPort {
  getScrape(): Promise<ScrapeSettings>;
  updateScrape(update: ScrapeSettingsUpdate): Promise<ScrapeSettings>;
}
