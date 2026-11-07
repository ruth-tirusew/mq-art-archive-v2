import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../../test/msw/server';
import { ApiError } from '$lib/adapters/api/client';
import { InvitesApi } from '$lib/adapters/api/invitesApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { toSaveDraftInput, fromQuestions } from '$lib/utils/questionnaireDraft';
import { fixtureRevision } from '../../../test/fixtures/questionnaire';

const base = PUBLIC_API_URL.replace(/\/$/, '');

describe('QuestionnaireApi (MSW)', () => {
	const api = new QuestionnaireApi();

	it('reads the draft and revisions', async () => {
		expect((await api.getDraft())?.status).toBe('draft');
		const revs = await api.listRevisions();
		expect(revs[0]?.number).toBe(1);
		expect((await api.getRevision('rev-1')).questions).toHaveLength(2);
	});

	it('returns null instead of throwing when there is no draft', async () => {
		server.use(http.get(`${base}/admin/v1/questionnaire/draft`, () => HttpResponse.json({ error: 'not found' }, { status: 404 })));
		expect(await api.getDraft()).toBeNull();
	});

	it('still throws for real errors', async () => {
		server.use(http.get(`${base}/admin/v1/questionnaire/draft`, () => HttpResponse.json({ error: 'boom' }, { status: 500 })));
		await expect(api.getDraft()).rejects.toMatchObject({ name: 'ApiError', status: 500 } satisfies Partial<ApiError>);
	});

	it('saves the whole draft as JSON, in order', async () => {
		const questions = fromQuestions(fixtureRevision.questions, new Set(['bio', 'income']));
		const saved = await api.saveDraft(toSaveDraftInput('minor', 'note', [questions[1], questions[0]]));
		expect(saved.questions.map((q) => q.key)).toEqual(['income', 'bio']);
		expect(saved.questions.map((q) => q.position)).toEqual([1, 2]);
	});

	it('sends the loaded draft’s timestamp and surfaces a stale-save refusal as a conflict', async () => {
		let sent: Record<string, unknown> = {};
		server.use(
			http.put(`${base}/admin/v1/questionnaire/draft`, async ({ request }) => {
				sent = (await request.json()) as Record<string, unknown>;
				return HttpResponse.json({ error: 'conflict: reload to see their changes' }, { status: 409 });
			})
		);
		const questions = fromQuestions(fixtureRevision.questions, new Set(['bio', 'income']));
		const input = { ...toSaveDraftInput('minor', '', questions), expected_updated_at: '2026-01-01T00:00:00Z' };
		await expect(api.saveDraft(input)).rejects.toMatchObject({ status: 409 });
		expect(sent.expected_updated_at).toBe('2026-01-01T00:00:00Z');
	});

	it('declares JSON on bodyless POSTs too, which forces a CORS preflight for cross-origin callers', async () => {
		let contentType: string | null = 'unset';
		server.use(
			http.post(`${base}/admin/v1/questionnaire/draft/publish`, ({ request }) => {
				contentType = request.headers.get('content-type');
				return HttpResponse.json(fixtureRevision);
			})
		);
		await api.publish();
		expect(contentType).toBe('application/json');
	});

	it('previews and sends a nudge', async () => {
		expect((await api.previewNudge())?.eligible).toBe(3);
		expect(await api.sendNudge()).toBe(3);
	});

	it('treats a missing nudge preview as null', async () => {
		server.use(http.get(`${base}/admin/v1/questionnaire/nudge`, () => HttpResponse.json({ error: 'not found' }, { status: 404 })));
		expect(await api.previewNudge()).toBeNull();
	});

	it('loads an application review view, or null when it has no answers', async () => {
		const view = await api.getApplicationAnswers('app-1');
		expect(view?.items[0]?.answer?.format).toBe('rich');
		expect(view?.items[1]?.answer).toBeNull();
		expect(await api.getApplicationAnswers('missing')).toBeNull();
	});

	it('discards the draft', async () => {
		await expect(api.discardDraft()).resolves.toBeUndefined();
	});
});

describe('InvitesApi (MSW)', () => {
	const api = new InvitesApi();

	it('lists a page of invites with the total', async () => {
		const page = await api.list('all', { limit: 50, offset: 0 });
		expect(page.items[0]?.status).toBe('live');
		expect(page.total).toBe(3);
	});

	it('sends the status filter and window', async () => {
		let search = '';
		server.use(
			http.get(`${base}/admin/v1/invites`, ({ request }) => {
				search = new URL(request.url).search;
				return HttpResponse.json([]);
			})
		);
		await api.list('expired', { limit: 25, offset: 50 });
		expect(search).toBe('?status=expired&limit=25&offset=50');
	});

	it('creates an invite and returns the one-time link', async () => {
		const created = await api.create({ email: 'new@example.com', handle: 'new_1' });
		expect(created.invite.email).toBe('new@example.com');
		expect(created.link).toContain('invite=');
	});

	it('surfaces conflicts as ApiError', async () => {
		await expect(api.create({ email: 'taken@example.com' })).rejects.toMatchObject({ status: 409 });
	});

	it('resends and revokes', async () => {
		expect((await api.resend('inv-1')).token).toBe('tok_abc');
		await expect(api.revoke('inv-1')).resolves.toBeUndefined();
	});
});
