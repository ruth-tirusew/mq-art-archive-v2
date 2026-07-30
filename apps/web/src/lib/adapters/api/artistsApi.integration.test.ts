import { describe, expect, it } from 'vitest';
import { apiFetch, ApiError } from '$lib/adapters/api/client';
import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { ArtApi } from '$lib/adapters/api/artApi';

describe('apiFetch (MSW)', () => {
	it('returns JSON on success', async () => {
		const data = await apiFetch<unknown[]>('/api/v1/artists');
		expect(Array.isArray(data)).toBe(true);
		expect(data.length).toBeGreaterThan(0);
	});

	it('throws ApiError with message from body', async () => {
		await expect(apiFetch('/api/v1/__error')).rejects.toMatchObject({
			name: 'ApiError',
			status: 500,
			message: 'boom'
		} satisfies Partial<ApiError>);
	});
});

describe('ArtistsApi (MSW)', () => {
	const api = new ArtistsApi();

	it('lists artists', async () => {
		const artists = await api.list();
		expect(artists[0]?.display_name).toBe('Selamawit Tesfaye');
	});

	it('gets artist by handle', async () => {
		const artist = await api.getByHandle('selam');
		expect(artist.slug).toBe('selamawit-tesfaye');
	});

	it('surfaces 404 as ApiError', async () => {
		await expect(api.getByHandle('missing')).rejects.toBeInstanceOf(ApiError);
	});
});

describe('ArtApi (MSW)', () => {
	const api = new ArtApi();

	it('lists posts', async () => {
		const posts = await api.list();
		expect(posts[0]?.title).toBe('Blue Market');
	});

	it('gets a post by id', async () => {
		const post = await api.getById('post-1');
		expect(post.artist_slug).toBe('selamawit-tesfaye');
	});
});
