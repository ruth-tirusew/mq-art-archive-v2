import { get, writable } from 'svelte/store';
import { AuthApi } from '$lib/adapters/api/authApi';
import { onUnauthorized } from '$lib/adapters/api/client';
import { currentUser } from '$lib/application/auth';

const api = new AuthApi();

/**
 * True when a signed-in user's session has ended mid-visit. The page is left as it is (a form the
 * artist is filling in must not be thrown away); a banner asks them to sign in again elsewhere.
 */
export const sessionExpired = writable(false);

// Paths where a 401 is an ordinary answer, not a sign the session ended: wrong credentials, the
// "who am I" probe itself, and invite signup.
const IGNORED = ['/api/v1/auth/login', '/api/v1/auth/register', '/api/v1/auth/me', '/api/v1/auth/logout', '/api/v1/invites/'];

let checking: Promise<void> | null = null;

/**
 * A 401 alone isn't proof: some endpoints answer 401 for a wrong current password. Ask the API
 * who we are; only if that also fails was the session really lost. Concurrent 401s share one check.
 */
export function checkSession(): Promise<void> {
	checking ??= api
		.me()
		.then((user) => {
			currentUser.set(user);
			sessionExpired.set(false);
		})
		.catch((err: unknown) => {
			if ((err as { status?: number }).status === 401) sessionExpired.set(true);
		})
		.finally(() => {
			checking = null;
		});
	return checking;
}

export function handleUnauthorized(path: string): void {
	if (get(currentUser) === null) return; // never signed in: a 401 is expected, not an expiry
	if (IGNORED.some((p) => path === p || (p.endsWith('/') && path.startsWith(p)))) return;
	void checkSession();
}

/** Wire the transport's 401 signal to the session check. Returns a function that unwires it. */
export function watchSession(): () => void {
	onUnauthorized(handleUnauthorized);
	return () => onUnauthorized(null);
}
