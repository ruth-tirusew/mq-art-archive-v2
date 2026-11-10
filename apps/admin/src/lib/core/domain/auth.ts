import { PUBLIC_API_URL } from '$env/static/public';

export interface User {
  id: string;
  email: string;
  role: string;
  display_name?: string;
  avatar_url?: string;
  has_password: boolean;
  google_linked?: boolean;
}

/**
 * Connect Google to the account that is already signed in. The user returns to `/settings` with
 * `?google=linked` or `?google=error&reason=…`.
 */
export function googleLinkUrl(origin?: string): string {
  const base =
    origin ??
    (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5174');
  return `${PUBLIC_API_URL}/api/v1/auth/google/link?return_to=${encodeURIComponent(`${base}/settings#account`)}`;
}

/** Build Google OAuth start URL. Pass `origin` during SSR (e.g. `$page.url.origin`). */
export function googleLoginUrl(origin?: string): string {
  const base =
    origin ??
    (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5174');
  const returnTo = `${base}/auth/callback`;
  return `${PUBLIC_API_URL}/api/v1/auth/google?return_to=${encodeURIComponent(returnTo)}`;
}
