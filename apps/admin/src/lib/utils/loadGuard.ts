import { redirect } from '@sveltejs/kit';
import { ApiError } from '$lib/adapters/api/client';

/** Run an admin loader, sending unauthenticated/forbidden responses to the login page. */
export async function requireAdmin<T>(load: () => Promise<T>): Promise<T> {
  try {
    return await load();
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect(302, '/login');
    }
    throw err;
  }
}
