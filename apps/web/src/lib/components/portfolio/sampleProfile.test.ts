import { describe, expect, it } from 'vitest';
import { postImageUrl } from '$lib/utils/display';
import { SAMPLE_ARTIST, SAMPLE_WORKS } from './sampleProfile';

describe('sample profile', () => {
	it('is a template, not a real person', () => {
		expect(SAMPLE_ARTIST.display_name).toBe('Your Name');
		expect(SAMPLE_ARTIST.social).toEqual({});
		expect(SAMPLE_ARTIST.contact?.email).toBeUndefined();
	});

	it('has three works that each draw their own image, with nothing to load', () => {
		expect(SAMPLE_WORKS).toHaveLength(3);
		for (const w of SAMPLE_WORKS) {
			const src = postImageUrl(w.media, w.id, w.artist_slug);
			expect(src).toMatch(/^data:image\/svg\+xml/);
			expect(w.artist_id).toBe(SAMPLE_ARTIST.id);
		}
		expect(new Set(SAMPLE_WORKS.map((w) => postImageUrl(w.media))).size).toBe(3);
	});
});
