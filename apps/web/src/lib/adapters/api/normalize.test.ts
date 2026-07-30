import { describe, expect, it } from 'vitest';
import { normalizeProfile, normalizeProfiles } from '$lib/adapters/api/normalize';
import { artistName, artistHandle } from '$lib/utils/fields';

describe('normalizeProfile', () => {
	it('maps snake_case API fields', () => {
		const profile = normalizeProfile({
			id: '1',
			slug: 'ada',
			display_name: 'Ada',
			handle: 'ada',
			status: 'approved'
		});
		expect(profile.display_name).toBe('Ada');
		expect(profile.slug).toBe('ada');
	});

	it('maps PascalCase API fields', () => {
		const profile = normalizeProfile({
			ID: '2',
			Slug: 'bekele',
			DisplayName: 'Bekele',
			Status: 'pending'
		});
		expect(profile.id).toBe('2');
		expect(profile.display_name).toBe('Bekele');
		expect(profile.status).toBe('pending');
	});

	it('returns empty array for non-array list payloads', () => {
		expect(normalizeProfiles(null)).toEqual([]);
		expect(normalizeProfiles({})).toEqual([]);
	});
});

describe('artist field helpers', () => {
	it('reads display_name and handle from profiles', () => {
		const artist = {
			id: '1',
			slug: 'selam',
			handle: 'selam',
			display_name: 'Selamawit'
		};
		expect(artistName(artist)).toBe('Selamawit');
		expect(artistHandle(artist)).toBe('selam');
	});
});
