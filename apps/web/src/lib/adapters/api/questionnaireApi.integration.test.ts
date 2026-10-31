import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../../test/msw/server';
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
