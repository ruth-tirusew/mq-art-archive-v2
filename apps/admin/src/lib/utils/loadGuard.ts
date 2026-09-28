import { error, redirect } from '@sveltejs/kit';
import { ApiError } from '$lib/adapters/api/client';

/**
 * Run an admin loader, sending unauthenticated/forbidden responses to the login page.
 * Any other failure (the API being unreachable, a 500 from the server, etc.) is turned
 * into a clean SvelteKit error() instead of an uncaught exception — a raw network
 * failure isn't an ApiError instance, so it used to bypass this guard entirely and
 * render as an unhandled 500 with no useful message.
 */
export async function requireAdmin<T>(load: () => Promise<T>): Promise<T> {
  try {
    return await load();
  } catch (err) {
    if (err instanceof ApiError) {
      if (err.status === 401 || err.status === 403) {
        redirect(302, '/login');
      }
      error(err.status, err.message);
    }
    error(503, 'Could not reach the API. Make sure the backend is running.');
  }
}
