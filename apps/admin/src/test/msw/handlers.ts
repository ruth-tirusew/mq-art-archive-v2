import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { fixtureArtists, fixtureArtist } from '../fixtures/artists';
import { fixturePosts, fixturePost } from '../fixtures/posts';
import { fixtureApplications, fixtureApplication } from '../fixtures/applications';

const base = PUBLIC_API_URL.replace(/\/$/, '');

export const handlers = [
	http.get(`${base}/admin/v1/artists`, () => HttpResponse.json(fixtureArtists)),
	http.get(`${base}/admin/v1/artists/:id`, ({ params }) => {
		const match = fixtureArtists.find((a) => a.id === params.id);
		if (!match) {
			return HttpResponse.json({ error: 'not found' }, { status: 404 });
		}
		return HttpResponse.json(match);
	}),
	http.patch(`${base}/admin/v1/artists/:id`, async ({ params, request }) => {
		const body = (await request.json()) as Partial<typeof fixtureArtist>;
		return HttpResponse.json({ ...fixtureArtist, id: params.id, ...body });
	}),
	http.post(`${base}/admin/v1/artists`, async ({ request }) => {
		const body = (await request.json()) as Partial<typeof fixtureArtist> & { display_name: string };
		return HttpResponse.json(
			{ ...fixtureArtist, ...body, id: 'artist-new', slug: body.slug ?? 'new-artist' },
			{ status: 201 }
		);
	}),
	http.put(`${base}/admin/v1/artists/:id`, async ({ params, request }) => {
		const body = (await request.json()) as Partial<typeof fixtureArtist>;
		return HttpResponse.json({ ...fixtureArtist, id: params.id, ...body });
	}),
	http.delete(`${base}/admin/v1/artists/:id`, () => new HttpResponse(null, { status: 204 })),
	http.get(`${base}/admin/v1/posts`, () => HttpResponse.json(fixturePosts)),
	http.get(`${base}/admin/v1/posts/:id`, ({ params }) => {
		if (params.id !== fixturePost.id) {
			return HttpResponse.json({ error: 'not found' }, { status: 404 });
		}
		return HttpResponse.json(fixturePost);
	}),
	http.post(`${base}/admin/v1/posts`, async ({ request }) => {
		const body = (await request.json()) as Partial<typeof fixturePost> & { artist_id: string; title: string };
		return HttpResponse.json(
			{ ...fixturePost, ...body, id: 'post-new', media: (body as { media_urls?: string[] }).media_urls?.map((url) => ({ url })) ?? fixturePost.media },
			{ status: 201 }
		);
	}),
	http.put(`${base}/admin/v1/posts/:id`, async ({ params, request }) => {
		const body = (await request.json()) as Partial<typeof fixturePost>;
		return HttpResponse.json({ ...fixturePost, id: params.id, ...body });
	}),
	http.patch(`${base}/admin/v1/posts/:id`, async ({ params, request }) => {
		const body = (await request.json()) as Partial<typeof fixturePost>;
		return HttpResponse.json({ ...fixturePost, id: params.id, ...body });
	}),
	http.delete(`${base}/admin/v1/posts/:id`, () => new HttpResponse(null, { status: 204 })),
	http.get(`${base}/admin/v1/applications`, () => HttpResponse.json(fixtureApplications)),
	http.get(`${base}/admin/v1/applications/:id`, ({ params }) => {
		if (params.id !== fixtureApplication.id) {
			return HttpResponse.json({ error: 'not found' }, { status: 404 });
		}
		return HttpResponse.json(fixtureApplication);
	}),
	http.get(`${base}/admin/v1/__error`, () =>
		HttpResponse.json({ error: 'boom' }, { status: 500 })
	)
];
