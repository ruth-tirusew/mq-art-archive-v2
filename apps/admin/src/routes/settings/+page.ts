import { error, redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { ApiError } from '$lib/adapters/api/client';
import { authService } from '$lib/application/auth';
import { settingsService } from '$lib/application/settings';

export const load: PageLoad = async () => {
  try {
    const user = await authService.load();
    if (!user) {
      redirect(303, '/login');
    }
    // Moderators use this page for their own profile, password and notifications. The scrape and Telegram
    // settings are admin-only on the API, so asking for them would only fail.
    const [notifications, scrape] = await Promise.all([
      authService.getNotifications(),
      user.role === 'admin' ? settingsService.getScrape() : Promise.resolve(null)
    ]);

    return { user, notifications, scrape };
  } catch (err) {
    if (err && typeof err === 'object' && 'status' in err && 'location' in err) {
      throw err;
    }
    if (err instanceof ApiError) {
      if (err.status === 401 || err.status === 403) {
        redirect(303, '/login');
      }
      error(err.status, err.message);
    }
    error(503, 'Could not reach the API. Make sure the backend is running.');
  }
};
