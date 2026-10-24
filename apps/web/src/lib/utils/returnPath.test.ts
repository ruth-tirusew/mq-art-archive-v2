import { describe, expect, it } from 'vitest';
import { safeReturnPath } from './returnPath';

describe('safeReturnPath', () => {
	it('keeps same-site relative paths', () => {
		expect(safeReturnPath('/studio')).toBe('/studio');
		expect(safeReturnPath('/wiki/some-entry?x=1#c')).toBe('/wiki/some-entry?x=1#c');
	});

	it.each([
		'https://evil.com',
		'http://evil.com/x',
		'//evil.com',
		'///evil.com',
		'/\\evil.com',
		'javascript:alert(1)',
		'studio',
		'/ok\nSet-Cookie: x=1',
		'/a\tb'
	])('rejects %j', (value) => {
		expect(safeReturnPath(value)).toBe('/studio');
	});

	it('falls back for empty values and honours a custom fallback', () => {
		expect(safeReturnPath(null)).toBe('/studio');
		expect(safeReturnPath('', '/account')).toBe('/account');
		expect(safeReturnPath('//x.io', '/account')).toBe('/account');
	});
});
