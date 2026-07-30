import { describe, expect, it } from 'vitest';
import { apiFetch, ApiError } from '$lib/adapters/api/client';
import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { PostsApi } from '$lib/adapters/api/postsApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';

describe('apiFetch (MSW)', () => {
	it('returns JSON on success', async () => {
		const data = await apiFetch<unknown[]>('/admin/v1/artists');
		expect(Array.isArray(data)).toBe(true);
		expect(data.length).toBeGreaterThan(0);
	});

	it('throws ApiError with message from body', async () => {
		await expect(apiFetch('/admin/v1/__error')).rejects.toMatchObject({
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

	it('gets artist by id', async () => {
		const artist = await api.getById('artist-1');
		expect(artist.slug).toBe('selamawit-tesfaye');
	});

	it('patches artist status', async () => {
		const artist = await api.patch('artist-1', { status: 'approved', featured: true });
		expect(artist.status).toBe('approved');
		expect(artist.featured).toBe(true);
	});
});

describe('PostsApi (MSW)', () => {
	const api = new PostsApi();

	it('lists posts', async () => {
		const posts = await api.list();
		expect(posts[0]?.title).toBe('Blue Market');
	});

	it('patches post status', async () => {
		const post = await api.patch('post-1', { status: 'archived' });
		expect(post.status).toBe('archived');
	});
});

describe('OnboardingApi (MSW)', () => {
	const api = new OnboardingApi();

	it('lists applications', async () => {
		const apps = await api.listPending();
		expect(apps[0]?.display_name).toBe('New Applicant');
	});
});
