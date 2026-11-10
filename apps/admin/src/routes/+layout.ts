import { redirect } from '@sveltejs/kit';
import { authService } from '$lib/application/auth';
import { adminGateRedirect, isPublicPath } from '$lib/utils/adminGate';
import type { LayoutLoad } from './$types';

// Runs before any admin page renders or loads (on the server for a first visit, in the browser after
// that), so a page can never mount, or start fetching, for someone who isn't a signed-in admin.
export const load: LayoutLoad = async ({ url }) => {
  if (isPublicPath(url.pathname)) return { user: null };
  const user = await authService.session();
  const to = adminGateRedirect(user);
  if (to) redirect(302, to);
  return { user };
};
