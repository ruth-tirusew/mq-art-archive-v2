export const LOGIN_PATH = '/login';
export const ACCESS_DENIED_PATH = '/login?error=access_denied';

/** Paths that must stay reachable without a session. `/accept-invite` is where an invited moderator picks a password. */
export function isPublicPath(pathname: string): boolean {
  return pathname === '/login' || pathname === '/accept-invite' || pathname.startsWith('/auth/');
}

/** Roles that may use the admin app. A moderator does the review and content work; see isAdminOnlyPath. */
export function isStaffRole(role: string): boolean {
  return role === 'admin' || role === 'moderator';
}

/** Where to send a visitor who may not see the admin, or null when they may. */
export function adminGateRedirect(user: { role: string } | null): string | null {
  if (!user) return LOGIN_PATH;
  return isStaffRole(user.role) ? null : ACCESS_DENIED_PATH;
}

/**
 * Pages only admins may open: user and role management. (The API enforces the same on its side; this just
 * keeps a moderator from landing on a page that could only show errors.) Settings stays open to everyone on
 * staff, because it holds each person's own profile and password; its integrations section is admin-only.
 */
export function isAdminOnlyPath(pathname: string): boolean {
  return pathname === '/users' || pathname.startsWith('/users/');
}

/** Where to send a moderator who opened an admin-only page, or null when they may be there. */
export function adminOnlyRedirect(user: { role: string }, pathname: string): string | null {
  return user.role !== 'admin' && isAdminOnlyPath(pathname) ? '/' : null;
}
