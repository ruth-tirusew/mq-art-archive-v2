import { PUBLIC_API_URL } from '$env/static/public';
import { safeReturnPath } from '$lib/utils/returnPath';

export interface User {
	id: string;
	email: string;
	role: string;
	display_name?: string;
	avatar_url?: string;
	has_password?: boolean;
	google_linked?: boolean;
	email_verified?: boolean;
}

export interface NotificationPreferences {
	email_on_new_application: boolean;
	email_on_event_sync_summary: boolean;
	newsletter_enabled: boolean;
}

export { safeReturnPath };

function currentOrigin(origin?: string): string {
	return origin ?? (typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173');
}

/** Build Google OAuth start URL. Safe during SSR when `window` is unavailable. */
export function googleLoginUrl(returnPath = '/studio', origin?: string): string {
	const callbackUrl = `${currentOrigin(origin)}/auth/callback?return_to=${encodeURIComponent(safeReturnPath(returnPath))}`;
	return `${PUBLIC_API_URL}/api/v1/auth/google?return_to=${encodeURIComponent(callbackUrl)}`;
}

/**
 * Start connecting Google to the account that is already signed in. The user comes back to
 * `returnPath` with `?google=linked` or `?google=error&reason=…`.
 */
export function googleLinkUrl(returnPath = '/account', origin?: string): string {
	const target = `${currentOrigin(origin)}${safeReturnPath(returnPath, '/account')}`;
	return `${PUBLIC_API_URL}/api/v1/auth/google/link?return_to=${encodeURIComponent(target)}`;
}
