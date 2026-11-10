import { describe, expect, it } from 'vitest';
import { createApplicantWorksService } from '$lib/application/applicantWorks';
import { workDraftProblem, type WorkDraft } from '$lib/core/domain/onboarding';
import type { ApplicantWorksPort } from '$lib/core/ports/applicantWorks';

const draft = (n: number, over: Partial<WorkDraft> = {}): WorkDraft => ({
	key: `k${n}`,
	file: new File(['x'], `piece-${n}.jpg`, { type: 'image/jpeg' }),
	previewUrl: `blob:local-${n}`,
	title: `Piece ${n}`,
	medium: '',
	year: '',
	status: 'pending',
	...over
});

function fakes(failOn?: string) {
	const log: string[] = [];
	const port: ApplicantWorksPort = {
		list: async () => [],
		add: async (input) => {
			log.push(`add:${input.title}:${input.year ?? '-'}`);
			return { id: `work-${input.title}`, media_asset_id: input.media_asset_id, url: 'u', title: input.title, position: 0 };
		},
		remove: async () => {},
		getPortrait: async () => null,
		setPortrait: async (id) => {
			log.push(`portrait:${id}`);
			return { media_asset_id: id, url: `https://img.test/${id}` };
		},
		removePortrait: async () => {}
	};
	const upload = async (file: File) => {
		log.push(`upload:${file.name}`);
		if (file.name === failOn) throw new Error('Cloudinary rejected the image upload.');
		return { secure_url: 'https://res.cloudinary.com/c/image/upload/x', id: `asset-${file.name}` };
	};
	return { service: createApplicantWorksService(port, upload), log };
}

describe('uploadDrafts: nothing is uploaded until this runs', () => {
	it('uploads then attaches each draft, in order, and marks it done', async () => {
		const { service, log } = fakes();
		const drafts = [draft(1), draft(2, { year: '2024' })];
		const progress: string[] = [];
		await service.uploadDrafts(drafts, { onProgress: (d, t) => progress.push(`${d}/${t}`) });
		expect(log).toEqual(['upload:piece-1.jpg', 'add:Piece 1:-', 'upload:piece-2.jpg', 'add:Piece 2:2024']);
		expect(drafts.map((d) => d.status)).toEqual(['done', 'done']);
		expect(drafts[0].workId).toBe('work-Piece 1');
		expect(progress).toEqual(['0/2', '1/2', '2/2']);
	});

	it('stops at the first failure and marks it, leaving later drafts untouched', async () => {
		const { service, log } = fakes('piece-2.jpg');
		const drafts = [draft(1), draft(2), draft(3)];
		await expect(service.uploadDrafts(drafts)).rejects.toThrow('Cloudinary rejected');
		expect(drafts.map((d) => d.status)).toEqual(['done', 'failed', 'pending']);
		expect(drafts[1].error).toContain('Cloudinary rejected');
		expect(log).not.toContain('upload:piece-3.jpg');
	});

	it('on retry only does what is left, so nothing is uploaded twice', async () => {
		const first = fakes('piece-2.jpg');
		const drafts = [draft(1), draft(2), draft(3)];
		await first.service.uploadDrafts(drafts).catch(() => {});

		const retry = fakes();
		await retry.service.uploadDrafts(drafts);
		expect(retry.log).toEqual(['upload:piece-2.jpg', 'add:Piece 2:-', 'upload:piece-3.jpg', 'add:Piece 3:-']);
		expect(drafts.every((d) => d.status === 'done')).toBe(true);
	});

	it('does nothing at all when there is nothing to upload', async () => {
		const { service, log } = fakes();
		await service.uploadDrafts([]);
		await service.uploadDrafts([draft(1, { status: 'done' })]);
		expect(log).toEqual([]);
	});
});

describe('workDraftProblem', () => {
	const now = new Date('2026-11-08');
	it('needs a title', () => {
		expect(workDraftProblem({ title: '  ', year: '' }, now)).toContain('title');
		expect(workDraftProblem({ title: 'Untitled', year: '' }, now)).toBeNull();
	});
	it('treats the year as optional but sane', () => {
		expect(workDraftProblem({ title: 'x', year: '2024' }, now)).toBeNull();
		expect(workDraftProblem({ title: 'x', year: '2027' }, now)).toBeNull();
		for (const bad of ['abc', '12', '2040', '20.5']) expect(workDraftProblem({ title: 'x', year: bad }, now)).toContain('Year');
	});
	it('caps the title length', () => {
		expect(workDraftProblem({ title: 'a'.repeat(201), year: '' }, now)).toContain('200');
	});
});

describe('uploadPortrait', () => {
	it('uploads the image then makes it the portrait', async () => {
		const { service, log } = fakes();
		const portrait = await service.uploadPortrait(new File(['x'], 'me.jpg', { type: 'image/jpeg' }));
		expect(log).toEqual(['upload:me.jpg', 'portrait:asset-me.jpg']);
		expect(portrait.media_asset_id).toBe('asset-me.jpg');
	});

	it('does not touch the portrait when the upload fails', async () => {
		const { service, log } = fakes('me.jpg');
		await expect(service.uploadPortrait(new File(['x'], 'me.jpg', { type: 'image/jpeg' }))).rejects.toThrow('Cloudinary rejected');
		expect(log).toEqual(['upload:me.jpg']);
	});
});
