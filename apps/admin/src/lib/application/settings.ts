import { SettingsApi } from '$lib/adapters/api/settingsApi';
import type { ScrapeSettings, ScrapeSettingsUpdate } from '$lib/core/domain/settings';

const api = new SettingsApi();

export const settingsService = {
  getScrape(): Promise<ScrapeSettings> {
    return api.getScrape();
  },

  updateScrape(update: ScrapeSettingsUpdate): Promise<ScrapeSettings> {
    return api.updateScrape(update);
  }
};
