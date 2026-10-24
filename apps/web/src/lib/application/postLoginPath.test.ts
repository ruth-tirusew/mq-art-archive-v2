import { describe, expect, it } from 'vitest';
import { postLoginPath } from './auth';
import type { User } from '$lib/core/domain/auth';

const user = (role: string) => ({ id: '1', email: 'a@b.c', role }) as User;

describe('postLoginPath', () => {
	it('sends artists to where they were going', () => {
		expect(postLoginPath(user('artist'), '/studio')).toBe('/studio');
		expect(postLoginPath(user('artist'), '/apply')).toBe('/apply');
	});

	it('sends readers home unless they were heading somewhere specific', () => {
		expect(postLoginPath(user('public'), '/studio')).toBe('/');
		expect(postLoginPath(user('public'), '/apply')).toBe('/apply');
		expect(postLoginPath(user('public'), '/wiki/x')).toBe('/wiki/x');
	});

	it('does not force contributors, admins or institutions into /apply', () => {
		expect(postLoginPath(user('contributor'), '/studio')).toBe('/studio');
		expect(postLoginPath(user('admin'), '/studio')).toBe('/');
		expect(postLoginPath(user('institution'), '/apply')).toBe('/');
	});
});
