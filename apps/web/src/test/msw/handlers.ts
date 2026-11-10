import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { fixtureArtists } from '../fixtures/artists';
import { fixturePosts, fixturePost } from '../fixtures/posts';
import { fixtureForm, fixtureNudge, fixturePublicAnswers } from '../fixtures/answers';

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
	http.get(`${base}/api/v1/artists/:slug/answers`, ({ params }) =>
		params.slug === 'selamawit-tesfaye'
			? HttpResponse.json(fixturePublicAnswers)
			: HttpResponse.json({ error: 'not found' }, { status: 404 })
	),
	// The API rejects any request with a body that isn't JSON (CSRF protection); the mocks do too.
	http.get(`${base}/api/v1/me/questionnaire`, () => HttpResponse.json(fixtureForm)),
	http.put(`${base}/api/v1/me/questionnaire/answers`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const body = (await request.json()) as { answers?: { key: string }[] };
		if (!body.answers?.length || body.answers.some((a) => a.key === 'nope')) {
			return HttpResponse.json({ error: 'validation failed: "nope" is not a question on the current form' }, { status: 400 });
		}
		return new HttpResponse(null, { status: 204 });
	}),
	http.get(`${base}/api/v1/me/questionnaire/nudge`, () => HttpResponse.json(fixtureNudge)),
	http.post(`${base}/api/v1/me/questionnaire/nudges/:revisionId/dismiss`, () => new HttpResponse(null, { status: 204 })),
	http.post(`${base}/api/v1/invites/preview`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const { token } = (await request.json()) as { token: string };
		return token === 'good'
			? HttpResponse.json({ email: 'ada@example.com', display_name: 'Ada', handle: 'ada_1', status: 'live' })
			: HttpResponse.json({ error: 'not found' }, { status: 404 });
	}),
	http.post(`${base}/api/v1/invites/register`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const body = (await request.json()) as { token: string; password: string; email?: string };
		if (body.token === 'taken') return HttpResponse.json({ error: 'conflict' }, { status: 409 });
		if (body.token !== 'good') return HttpResponse.json({ error: 'not found' }, { status: 404 });
		if (body.password.length < 8) {
			return HttpResponse.json({ error: 'validation failed: password must be at least 8 characters' }, { status: 400 });
		}
		// like the real API: the account's email is the invite's, whatever the body says
		return HttpResponse.json({ id: 'u-1', email: 'ada@example.com', role: 'public', email_verified: true }, { status: 201 });
	}),
	http.post(`${base}/api/v1/auth/verify-email`, async ({ request }) => {
		const body = (await request.json()) as { token: string };
		return body.token === 'good'
			? new HttpResponse(null, { status: 204 })
			: HttpResponse.json({ error: 'invalid or expired token' }, { status: 400 });
	}),
	http.post(`${base}/api/v1/applications/invite`, async ({ request }) => {
		const body = (await request.json()) as { token: string };
		return body.token === 'good'
			? HttpResponse.json({ id: 'app-1', applicant_type: 'artist', display_name: 'Ada', status: 'pending', invite_id: 'inv-1' }, { status: 201 })
			: HttpResponse.json({ error: 'conflict' }, { status: 409 });
	}),
	http.get(`${base}/api/v1/me/application/works`, () =>
		HttpResponse.json([{ id: 'w-1', media_asset_id: 'a-1', url: 'https://res.cloudinary.com/c/image/upload/mq/a.jpg', blurhash: 'LEHV6nWB2yk8pyo0adR*.7kCMdnj', title: 'First', position: 0 }])
	),
	http.post(`${base}/api/v1/me/application/works`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const body = (await request.json()) as { media_asset_id: string; title: string };
		if (body.media_asset_id === 'locked') {
			return HttpResponse.json({ error: 'your application is being reviewed' }, { status: 409 });
		}
		return HttpResponse.json({ id: 'w-2', media_asset_id: body.media_asset_id, url: 'https://res.cloudinary.com/c/image/upload/mq/b.jpg', title: body.title, position: 1 }, { status: 201 });
	}),
	http.delete(`${base}/api/v1/me/application/works/:id`, ({ params }) =>
		params.id === 'missing' ? HttpResponse.json({ error: 'not found' }, { status: 404 }) : new HttpResponse(null, { status: 204 })
	),
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
