import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { get } from 'svelte/store';
import { apiFetch } from '$lib/adapters/api/client';
import { currentUser } from '$lib/application/auth';
import { checkSession, sessionExpired, watchSession } from '$lib/application/session';
import type { User } from '$lib/core/domain/auth';

const ada = { id: 'u1', email: 'ada@example.com', role: 'artist' } as User;

/** Answers each path with the given status; /auth/me answers `meStatus`. */
function stubApi(statusFor: Record<string, number>, meStatus: number) {
	const fetchMock = vi.fn(async (url: string) => {
		const path = new URL(url, 'http://api.test').pathname;
		const status = path === '/api/v1/auth/me' ? meStatus : (statusFor[path] ?? 200);
		const body = status === 200 ? ada : { error: 'unauthorized' };
		return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
	});
	vi.stubGlobal('fetch', fetchMock);
	return fetchMock;
}

const meCalls = (f: ReturnType<typeof stubApi>) => f.mock.calls.filter(([u]) => String(u).endsWith('/api/v1/auth/me')).length;

let unwatch: () => void;
beforeEach(() => {
	sessionExpired.set(false);
	unwatch = watchSession();
});
afterEach(() => {
	unwatch();
	currentUser.set(null);
	vi.unstubAllGlobals();
});

describe('session expiry', () => {
	it('shows the banner when a signed-in user is really signed out', async () => {
		currentUser.set(ada);
		stubApi({ '/api/v1/me/profile': 401 }, 401);
		await expect(apiFetch('/api/v1/me/profile', { method: 'PUT' })).rejects.toThrow();
		await checkSession(); // the check started by the 401; wait for it
		expect(get(sessionExpired)).toBe(true);
	});

	it('ignores a 401 that only means a wrong current password', async () => {
		currentUser.set(ada);
		const f = stubApi({ '/api/v1/auth/me/email': 401 }, 200);
		await expect(apiFetch('/api/v1/auth/me/email', { method: 'PUT' })).rejects.toThrow();
		await checkSession();
		expect(get(sessionExpired)).toBe(false);
		expect(meCalls(f)).toBeGreaterThan(0);
	});

	it('never treats a 401 as an expiry for someone who was not signed in', async () => {
		const f = stubApi({ '/api/v1/me/favorites': 401 }, 401);
		await expect(apiFetch('/api/v1/me/favorites')).rejects.toThrow();
		expect(meCalls(f)).toBe(0);
		expect(get(sessionExpired)).toBe(false);
	});

	it('does not react to a failed sign-in', async () => {
		currentUser.set(ada);
		const f = stubApi({ '/api/v1/auth/login': 401 }, 401);
		await expect(apiFetch('/api/v1/auth/login', { method: 'POST' })).rejects.toThrow();
		expect(meCalls(f)).toBe(0);
		expect(get(sessionExpired)).toBe(false);
	});

	it('clears the banner once they have signed in again', async () => {
		currentUser.set(ada);
		sessionExpired.set(true);
		stubApi({}, 200);
		await checkSession();
		expect(get(sessionExpired)).toBe(false);
	});
});
