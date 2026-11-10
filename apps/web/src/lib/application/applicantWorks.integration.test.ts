import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../test/msw/server';
import { applicantWorksService } from '$lib/composition';
import { MIN_APPLICATION_WORKS, worksShortfall } from '$lib/core/domain/onboarding';
import { isErrorKind } from '$lib/core/domain/errors';

const base = PUBLIC_API_URL.replace(/\/$/, '');

describe('worksShortfall', () => {
	it('counts down to the five an artist needs, then stops', () => {
		expect(MIN_APPLICATION_WORKS).toBe(5);
		expect([0, 1, 4, 5, 9].map(worksShortfall)).toEqual([5, 4, 1, 0, 0]);
	});
});

describe('applicantWorksService (MSW)', () => {
	it('lists the applicant’s works with their blur hashes', async () => {
		const works = await applicantWorksService.list();
		expect(works).toHaveLength(1);
		expect(works[0]).toMatchObject({ title: 'First', blurhash: 'LEHV6nWB2yk8pyo0adR*.7kCMdnj' });
	});

	it('adds a work as JSON, trimming what was typed and dropping blank optional fields', async () => {
		let sent: Record<string, unknown> = {};
		server.use(
			http.post(`${base}/api/v1/me/application/works`, async ({ request }) => {
				sent = (await request.json()) as Record<string, unknown>;
				return HttpResponse.json({ id: 'w-9', media_asset_id: 'a-9', url: 'u', title: 'x', position: 0 }, { status: 201 });
			})
		);
		await applicantWorksService.add({ media_asset_id: 'a-9', title: '  Untitled  ', medium: '   ', year: 2024 });
		expect(sent).toEqual({ media_asset_id: 'a-9', title: 'Untitled', year: 2024 });
	});

	it('refuses a blank title without calling the server', async () => {
		let called = false;
		server.use(http.post(`${base}/api/v1/me/application/works`, () => ((called = true), HttpResponse.json({}, { status: 201 }))));
		await expect(applicantWorksService.add({ media_asset_id: 'a-1', title: '   ' })).rejects.toThrow('title');
		expect(called).toBe(false);
	});

	it('surfaces the lock while the application is under review as a conflict', async () => {
		const failure = await applicantWorksService.add({ media_asset_id: 'locked', title: 'Late' }).catch((e) => e);
		expect(isErrorKind(failure, 'conflict')).toBe(true);
	});

	it('removes a work, and says not-found for one that isn’t theirs', async () => {
		await expect(applicantWorksService.remove('w-1')).resolves.toBeUndefined();
		const failure = await applicantWorksService.remove('missing').catch((e) => e);
		expect(isErrorKind(failure, 'not_found')).toBe(true);
	});
});
