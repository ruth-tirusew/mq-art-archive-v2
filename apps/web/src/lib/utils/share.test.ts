import { describe, expect, it } from 'vitest';
import { profileUrl, shareDescription, shareImage } from './share';

describe('share helpers', () => {
	it('builds the short profile address from the site origin', () => {
		expect(profileUrl('https://artiv.et', 'ada')).toBe('https://artiv.et/@ada');
		expect(profileUrl('http://localhost:8082/', 'ada')).toBe('http://localhost:8082/@ada');
	});

	it('only offers real, fetchable pictures to link previews', () => {
		expect(shareImage('https://res.cloudinary.com/c/image/upload/a.jpg')).toBe('https://res.cloudinary.com/c/image/upload/a.jpg');
		expect(shareImage('http://example.com/a.png')).toBe('http://example.com/a.png');
		for (const bad of ['', undefined, null, 'data:image/svg+xml;utf8,<svg/>', 'blob:http://x/1', '/relative/a.jpg', 'javascript:alert(1)']) {
			expect(shareImage(bad as string | undefined)).toBeUndefined();
		}
	});

	it('uses the tagline, else a trimmed start of the bio', () => {
		expect(shareDescription('Light and memory', 'A long bio')).toBe('Light and memory');
		expect(shareDescription('  ', 'A short bio.')).toBe('A short bio.');
		const long = 'word '.repeat(80);
		const out = shareDescription(undefined, long, 40);
		expect(out.length).toBeLessThanOrEqual(41);
		expect(out.endsWith('…')).toBe(true);
		expect(shareDescription(undefined, undefined)).toBe('');
	});
});
