import { apiFetch } from '$lib/adapters/api/client';
import type { ScrapeSettings, ScrapeSettingsUpdate } from '$lib/core/domain/settings';
import type { SettingsPort } from '$lib/core/ports/settings';

export class SettingsApi implements SettingsPort {
  getScrape(): Promise<ScrapeSettings> {
    return apiFetch<ScrapeSettings>('/admin/v1/settings/scrape');
  }

  updateScrape(update: ScrapeSettingsUpdate): Promise<ScrapeSettings> {
    return apiFetch<ScrapeSettings>('/admin/v1/settings/scrape', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(update)
    });
  }
}
