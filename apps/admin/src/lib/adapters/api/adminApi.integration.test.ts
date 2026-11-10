import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../../test/msw/server';
import { apiFetch, ApiError } from '$lib/adapters/api/client';
import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { PostsApi } from '$lib/adapters/api/postsApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';

const base = PUBLIC_API_URL.replace(/\/$/, '');

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
		const page = await api.list('pending', { limit: 50, offset: 0 });
		expect(page.items[0]?.display_name).toBe('New Applicant');
		// Without X-Total-Count the page itself is the whole list.
		expect(page.total).toBe(page.items.length);
	});

	it('sends the status and paging, and reads the total from X-Total-Count', async () => {
		let search = '';
		server.use(
			http.get(`${base}/admin/v1/applications`, ({ request }) => {
				search = new URL(request.url).search;
				return HttpResponse.json([{ id: 'a-1', display_name: 'Ada', applicant_type: 'artist', status: 'approved' }], {
					headers: { 'X-Total-Count': '120' }
				});
			})
		);
		const page = await api.list('approved', { limit: 50, offset: 100 });
		expect(search).toBe('?status=approved&limit=50&offset=100');
		expect(page.total).toBe(120);
		expect(page.items).toHaveLength(1);
	});
});
