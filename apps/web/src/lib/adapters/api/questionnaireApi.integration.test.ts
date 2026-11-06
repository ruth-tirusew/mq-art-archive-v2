import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../../test/msw/server';
import { InviteApi } from '$lib/adapters/api/inviteApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { questionnaireService } from '$lib/application/questionnaire';

const base = PUBLIC_API_URL.replace(/\/$/, '');

describe('QuestionnaireApi (MSW)', () => {
	const api = new QuestionnaireApi();

	it('lists an artist’s public answers in order', async () => {
		const answers = await api.listArtistAnswers('selamawit-tesfaye');
		expect(answers.map((a) => a.key)).toEqual(['bio', 'city', 'site']);
		expect(answers[0]).toMatchObject({ format: 'rich', lang: 'am' });
	});

	it('throws ApiError for an unknown or non-public artist', async () => {
		await expect(api.listArtistAnswers('nobody')).rejects.toMatchObject({ name: 'ApiError', status: 404 });
	});

	it('encodes the slug into the path', async () => {
		let requested = '';
		server.use(
			http.get(`${base}/api/v1/artists/:slug/answers`, ({ request }) => {
				requested = new URL(request.url).pathname;
				return HttpResponse.json([]);
			})
		);
		await api.listArtistAnswers('a/b?c');
		expect(requested).toBe('/api/v1/artists/a%2Fb%3Fc/answers');
	});
});

describe('questionnaireService (MSW)', () => {
	it('returns answers when the API has them', async () => {
		expect(await questionnaireService.listArtistAnswers('selamawit-tesfaye')).toHaveLength(3);
	});

	it('shows no Q&A instead of failing the page when the artist has none or the API errors', async () => {
		expect(await questionnaireService.listArtistAnswers('nobody')).toEqual([]);
		server.use(http.get(`${base}/api/v1/artists/:slug/answers`, () => HttpResponse.json({ error: 'boom' }, { status: 500 })));
		expect(await questionnaireService.listArtistAnswers('selamawit-tesfaye')).toEqual([]);
		server.use(http.get(`${base}/api/v1/artists/:slug/answers`, () => HttpResponse.error()));
		expect(await questionnaireService.listArtistAnswers('selamawit-tesfaye')).toEqual([]);
	});
});

describe('QuestionnaireApi: applicant side (MSW)', () => {
	const api = new QuestionnaireApi();

	it('loads the form with the saved answers', async () => {
		const form = await api.getForm();
		expect(form?.revision.questions).toHaveLength(2);
		expect(form?.answers[0]).toMatchObject({ key: 'bio', format: 'rich' });
	});

	it('treats "no questionnaire published yet" as null, so the plain application is used', async () => {
		server.use(http.get(`${base}/api/v1/me/questionnaire`, () => HttpResponse.json({ error: 'not found' }, { status: 404 })));
		expect(await api.getForm()).toBeNull();
	});

	it('still throws for real failures, so a broken API is not mistaken for "no questionnaire"', async () => {
		server.use(http.get(`${base}/api/v1/me/questionnaire`, () => HttpResponse.json({ error: 'boom' }, { status: 500 })));
		await expect(api.getForm()).rejects.toMatchObject({ status: 500 });
	});

	it('saves answers as JSON', async () => {
		await expect(api.saveAnswers([{ key: 'bio', body: '<p>x</p>', format: 'rich', lang: 'en' }])).resolves.toBeUndefined();
	});

	it('can send a save with keepalive, for when the page is closing', async () => {
		let keepalive: boolean | undefined;
		server.use(
			http.put(`${base}/api/v1/me/questionnaire/answers`, ({ request }) => {
				keepalive = request.keepalive;
				return new HttpResponse(null, { status: 204 });
			})
		);
		await api.saveAnswers([{ key: 'bio', body: 'x', format: 'plain', lang: 'en' }], { keepalive: true });
		expect(keepalive).toBe(true);
	});

	it('surfaces the server’s validation message', async () => {
		await expect(api.saveAnswers([{ key: 'nope', body: 'x', format: 'plain', lang: 'en' }])).rejects.toMatchObject({
			status: 400,
			message: expect.stringContaining('not a question')
		});
	});

	it('reads and dismisses a nudge; no nudge is null', async () => {
		expect((await api.getNudge())?.questions[0]?.key).toBe('influences');
		await expect(api.dismissNudge('rev-2')).resolves.toBeUndefined();
		server.use(http.get(`${base}/api/v1/me/questionnaire/nudge`, () => HttpResponse.json({ error: 'not found' }, { status: 404 })));
		expect(await api.getNudge()).toBeNull();
	});

	it('the service swallows nudge failures: a nudge must never get in the way', async () => {
		server.use(http.get(`${base}/api/v1/me/questionnaire/nudge`, () => HttpResponse.json({ error: 'boom' }, { status: 500 })));
		expect(await questionnaireService.getNudge()).toBeNull();
	});
});

describe('InviteApi (MSW)', () => {
	const api = new InviteApi();

	it('previews an invite, with the token in the body', async () => {
		expect(await api.preview('good')).toMatchObject({ email: 'ada@example.com', status: 'live', handle: 'ada_1' });
		await expect(api.preview('bad')).rejects.toMatchObject({ status: 404 });
	});

	it('never puts the token in the URL', async () => {
		let url = '';
		server.use(
			http.post(`${base}/api/v1/invites/preview`, ({ request }) => {
				url = request.url;
				return HttpResponse.json({ email: 'a@b.co', status: 'live' });
			})
		);
		await api.preview('secret-token-value');
		expect(url).not.toContain('secret-token-value');
	});

	it('registers an account from the invite with just a token and a password', async () => {
		const user = await api.register('good', 'correct-horse-9');
		expect(user).toMatchObject({ email: 'ada@example.com', email_verified: true, role: 'public' });
	});

	it('sends only a token and a password: the address is the invite’s, not the caller’s', async () => {
		let body: Record<string, unknown> = {};
		server.use(
			http.post(`${base}/api/v1/invites/register`, async ({ request }) => {
				body = (await request.json()) as Record<string, unknown>;
				return HttpResponse.json({ id: 'u', email: 'a@b.co', role: 'public', email_verified: true }, { status: 201 });
			})
		);
		await api.register('good', 'correct-horse-9');
		expect(Object.keys(body).sort()).toEqual(['password', 'token']);
	});

	it('surfaces registration refusals', async () => {
		await expect(api.register('taken', 'correct-horse-9')).rejects.toMatchObject({ status: 409 });
		await expect(api.register('nope', 'correct-horse-9')).rejects.toMatchObject({ status: 404 });
		await expect(api.register('good', 'short')).rejects.toMatchObject({
			status: 400,
			message: expect.stringContaining('at least 8')
		});
	});

	it('claims an invite and returns the pending application', async () => {
		const app = await api.claim({ token: 'good', display_name: 'Ada', requested_handle: 'ada_1' });
		expect(app).toMatchObject({ status: 'pending', invite_id: 'inv-1' });
		await expect(api.claim({ token: 'used' })).rejects.toMatchObject({ status: 409 });
	});

	it('sends the applicant’s notes with the claim', async () => {
		let sent: Record<string, unknown> = {};
		server.use(
			http.post(`${base}/api/v1/applications/invite`, async ({ request }) => {
				sent = (await request.json()) as Record<string, unknown>;
				return HttpResponse.json({ id: 'app-1', status: 'pending' }, { status: 201 });
			})
		);
		await api.claim({ token: 'good', requested_handle: 'ada_1', notes: 'Instagram: @ada' });
		expect(sent).toMatchObject({ token: 'good', notes: 'Instagram: @ada' });
	});
});
