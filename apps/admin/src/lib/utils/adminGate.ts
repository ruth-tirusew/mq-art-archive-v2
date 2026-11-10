export const LOGIN_PATH = '/login';
export const ACCESS_DENIED_PATH = '/login?error=access_denied';

/** Paths that must stay reachable without a session. */
export function isPublicPath(pathname: string): boolean {
  return pathname === '/login' || pathname.startsWith('/auth/');
}

/** Where to send a visitor who may not see the admin, or null when they may. */
export function adminGateRedirect(user: { role: string } | null): string | null {
  if (!user) return LOGIN_PATH;
  return user.role === 'admin' ? null : ACCESS_DENIED_PATH;
}
