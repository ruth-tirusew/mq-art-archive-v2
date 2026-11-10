import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { fixtureArtists, fixtureArtist } from '../fixtures/artists';
import { fixturePosts, fixturePost } from '../fixtures/posts';
import { fixtureApplications, fixtureApplication } from '../fixtures/applications';
import {
	fixtureDraft,
	fixtureInvite,
	fixtureInviteCreated,
	fixtureNudgePreview,
	fixtureReviewView,
	fixtureRevision
} from '../fixtures/questionnaire';

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
	// Questionnaire. Like the real API, every request with a body must declare application/json.
	http.get(`${base}/admin/v1/questionnaire/draft`, () => HttpResponse.json(fixtureDraft)),
	http.post(`${base}/admin/v1/questionnaire/draft`, () => HttpResponse.json(fixtureDraft)),
	http.put(`${base}/admin/v1/questionnaire/draft`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const body = (await request.json()) as { questions: { key: string }[] };
		return HttpResponse.json({
			...fixtureDraft,
			questions: body.questions.map((q, i) => ({ ...q, position: i + 1 }))
		});
	}),
	http.delete(`${base}/admin/v1/questionnaire/draft`, () => new HttpResponse(null, { status: 204 })),
	http.post(`${base}/admin/v1/questionnaire/draft/publish`, () => HttpResponse.json(fixtureRevision)),
	http.get(`${base}/admin/v1/questionnaire/revisions`, () =>
		HttpResponse.json([{ ...fixtureRevision, questions: [] }])
	),
	http.get(`${base}/admin/v1/questionnaire/revisions/:id`, ({ params }) =>
		params.id === fixtureRevision.id
			? HttpResponse.json(fixtureRevision)
			: HttpResponse.json({ error: 'not found' }, { status: 404 })
	),
	http.get(`${base}/admin/v1/questionnaire/nudge`, () => HttpResponse.json(fixtureNudgePreview)),
	http.post(`${base}/admin/v1/questionnaire/nudge`, () => HttpResponse.json({ sent: 3 })),
	http.get(`${base}/admin/v1/applications/:id/answers`, ({ params }) =>
		params.id === fixtureApplication.id
			? HttpResponse.json(fixtureReviewView)
			: HttpResponse.json({ error: 'not found' }, { status: 404 })
	),
	http.get(`${base}/admin/v1/invites`, () => HttpResponse.json([fixtureInvite], { headers: { 'X-Total-Count': '3' } })),
	http.post(`${base}/admin/v1/invites`, async ({ request }) => {
		if (!request.headers.get('content-type')?.startsWith('application/json')) {
			return HttpResponse.json({ error: 'content type must be application/json' }, { status: 415 });
		}
		const body = (await request.json()) as { email: string };
		if (body.email === 'taken@example.com') {
			return HttpResponse.json({ error: 'conflict' }, { status: 409 });
		}
		return HttpResponse.json(
			{ ...fixtureInviteCreated, invite: { ...fixtureInvite, email: body.email } },
			{ status: 201 }
		);
	}),
	http.post(`${base}/admin/v1/invites/:id/resend`, () => HttpResponse.json(fixtureInviteCreated, { status: 201 })),
	http.delete(`${base}/admin/v1/invites/:id`, () => new HttpResponse(null, { status: 204 })),
	http.get(`${base}/admin/v1/__error`, () =>
		HttpResponse.json({ error: 'boom' }, { status: 500 })
	)
];
