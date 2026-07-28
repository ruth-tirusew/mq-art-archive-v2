import { PUBLIC_API_URL } from '$env/static/public';

export interface User {
  id: string;
  email: string;
  role: string;
  display_name?: string;
  avatar_url?: string;
  has_password: boolean;
}

/** Build Google OAuth start URL. Pass `origin` during SSR (e.g. `$page.url.origin`). */
export function googleLoginUrl(origin?: string): string {
  const base =
    origin ??
    (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5174');
  const returnTo = `${base}/auth/callback`;
  return `${PUBLIC_API_URL}/api/v1/auth/google?return_to=${encodeURIComponent(returnTo)}`;
}
