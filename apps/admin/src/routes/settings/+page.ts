import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { ApiError } from '$lib/adapters/api/client';
import { authService } from '$lib/application/auth';
import { settingsService } from '$lib/application/settings';

export const load: PageLoad = async () => {
  try {
    const [user, notifications, scrape] = await Promise.all([
      authService.load(),
      authService.getNotifications(),
      settingsService.getScrape()
    ]);

    if (!user) {
      throw redirect(303, '/login');
    }

    return { user, notifications, scrape };
  } catch (err) {
    if (err && typeof err === 'object' && 'status' in err && 'location' in err) {
      throw err;
    }
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      throw redirect(303, '/login');
    }
    throw err;
  }
};
