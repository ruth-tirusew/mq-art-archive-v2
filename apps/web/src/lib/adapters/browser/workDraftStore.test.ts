import { describe, expect, it } from 'vitest';
import type { WorkDraft } from '$lib/core/domain/onboarding';
import { fromStored, isFresh, toStored, workDraftStore } from './workDraftStore';

const file = (name: string) => new File([new Uint8Array([1, 2, 3])], name, { type: 'image/png' });

const draft = (over: Partial<WorkDraft> = {}): WorkDraft => ({
	key: 'k1',
	file: file('sunrise.png'),
	previewUrl: 'blob:old',
	title: 'Sunrise',
	medium: 'Oil',
	year: '2024',
	status: 'pending',
	...over
});

describe('work draft storage', () => {
	it('round-trips what the applicant typed and the image itself', () => {
		const [restored] = fromStored(toStored([draft()]), () => 'blob:new');
		expect(restored).toMatchObject({ key: 'k1', title: 'Sunrise', medium: 'Oil', year: '2024', status: 'pending', previewUrl: 'blob:new' });
		expect(restored.file.name).toBe('sunrise.png');
		expect(restored.file.type).toBe('image/png');
		expect(restored.file.size).toBe(3);
	});

	it('rebuilds a File from the Blob IndexedDB hands back', () => {
		const stored = toStored([draft()]).map((s) => ({ ...s, file: new Blob([s.file], { type: '' }) }));
		const [restored] = fromStored(stored, () => 'blob:x');
		expect(restored.file).toBeInstanceOf(File);
		expect(restored.file.name).toBe('sunrise.png');
		expect(restored.file.type).toBe('image/png');
	});

	it('restores a piece that already reached the server as done, so it is not uploaded twice', () => {
		const [restored] = fromStored(toStored([draft({ status: 'done', workId: 'w-9' })]), () => 'blob:x');
		expect(restored.status).toBe('done');
		expect(restored.workId).toBe('w-9');
	});

	it('restores a piece that was mid-upload as pending', () => {
		const [restored] = fromStored(toStored([draft({ status: 'uploading' })]), () => 'blob:x');
		expect(restored.status).toBe('pending');
	});

	it('never stores the preview URL, which dies with the page', () => {
		expect(toStored([draft()])[0]).not.toHaveProperty('previewUrl');
	});

	it('drops drafts left untouched for over a month', () => {
		const now = Date.UTC(2026, 10, 10);
		expect(isFresh(now - 29 * 86_400_000, now)).toBe(true);
		expect(isFresh(now - 31 * 86_400_000, now)).toBe(false);
	});

	it('is harmless where IndexedDB does not exist', async () => {
		// Node has no IndexedDB: the form must carry on as if nothing was stored.
		await expect(workDraftStore.load('u1')).resolves.toEqual([]);
		await expect(workDraftStore.save('u1', [draft()])).resolves.toBe(false);
		await expect(workDraftStore.clear('u1')).resolves.toBeUndefined();
	});
});
