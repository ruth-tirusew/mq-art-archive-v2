import { PUBLIC_API_URL } from '$env/static/public';

export interface User {
	id: string;
	email: string;
	role: string;
	display_name?: string;
	avatar_url?: string;
	has_password?: boolean;
	email_verified?: boolean;
}

export interface NotificationPreferences {
	email_on_new_application: boolean;
	email_on_event_sync_summary: boolean;
	newsletter_enabled: boolean;
}

/** Build Google OAuth start URL. Safe during SSR when `window` is unavailable. */
export function googleLoginUrl(returnPath = '/studio', origin?: string): string {
	const base =
		origin ??
		(typeof window !== 'undefined' ? window.location.origin : 'http://localhost:5173');
	const callbackUrl = `${base}/auth/callback?return_to=${encodeURIComponent(returnPath)}`;
	return `${PUBLIC_API_URL}/api/v1/auth/google?return_to=${encodeURIComponent(callbackUrl)}`;
}
