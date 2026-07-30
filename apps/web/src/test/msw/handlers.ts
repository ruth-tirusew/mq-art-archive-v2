import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { fixtureArtists } from '../fixtures/artists';
import { fixturePosts, fixturePost } from '../fixtures/posts';

const base = PUBLIC_API_URL.replace(/\/$/, '');

export const handlers = [
	http.get(`${base}/api/v1/artists`, () => HttpResponse.json(fixtureArtists)),
	http.get(`${base}/api/v1/profiles/@:handle`, ({ params }) => {
		const match = fixtureArtists.find((a) => a.handle === params.handle);
		if (!match) {
			return HttpResponse.json({ error: 'not found' }, { status: 404 });
		}
		return HttpResponse.json(match);
	}),
	http.get(`${base}/api/v1/posts`, () => HttpResponse.json(fixturePosts)),
	http.get(`${base}/api/v1/posts/:id`, ({ params }) => {
		if (params.id !== fixturePost.id) {
			return HttpResponse.json({ error: 'not found' }, { status: 404 });
		}
		return HttpResponse.json(fixturePost);
	}),
	http.get(`${base}/api/v1/artists/:slug/posts`, () => HttpResponse.json(fixturePosts)),
	http.get(`${base}/api/v1/handles/:handle/available`, ({ params }) =>
		HttpResponse.json({ handle: params.handle, available: params.handle !== 'taken' })
	),
	http.get(`${base}/api/v1/search`, ({ request }) => {
		const q = new URL(request.url).searchParams.get('q');
		return HttpResponse.json({
			artists: q ? fixtureArtists.slice(0, 1) : [],
			posts: q ? fixturePosts.slice(0, 1) : [],
			articles: q ? [{ id: 'article-1', slug: 'ceramics', title: 'Ceramics', body: 'Body', status: 'published' }] : [],
			events: q ? [{ id: 'event-1', slug: 'opening', title: 'Opening', starts_at: '2026-08-01T10:00:00Z', status: 'approved' }] : []
		});
	}),
	http.get(`${base}/api/v1/__error`, () =>
		HttpResponse.json({ error: 'boom' }, { status: 500 })
	)
];
