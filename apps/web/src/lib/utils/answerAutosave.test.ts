import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { AnswerInput } from '$lib/core/domain/questionnaire';
import { AnswerAutosave, type SaveStatus } from './answerAutosave';

const a = (key: string, body: string): AnswerInput => ({ key, body, format: 'plain', lang: 'en' });

function setup(save = vi.fn().mockResolvedValue(undefined)) {
	const statuses: SaveStatus[] = [];
	const auto = new AnswerAutosave({ save, delayMs: 800, onStatus: (s) => statuses.push(s) });
	return { auto, save, statuses };
}

beforeEach(() => vi.useFakeTimers());
afterEach(() => vi.useRealTimers());

describe('AnswerAutosave', () => {
	it('waits for a quiet period and saves once, keeping only the latest value per key', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'h'));
		await vi.advanceTimersByTimeAsync(500);
		auto.set(a('bio', 'he'));
		auto.set(a('city', 'Addis'));
		await vi.advanceTimersByTimeAsync(700);
		expect(save).not.toHaveBeenCalled(); // the timer restarted with each edit
		await vi.advanceTimersByTimeAsync(200);
		expect(save).toHaveBeenCalledTimes(1);
		expect(save.mock.calls[0][0]).toEqual([a('bio', 'he'), a('city', 'Addis')]);
	});

	it('sends only what changed since the last save', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'one'));
		await vi.advanceTimersByTimeAsync(800);
		auto.set(a('city', 'Addis'));
		await vi.advanceTimersByTimeAsync(800);
		expect(save.mock.calls.map((c) => c[0].map((x: AnswerInput) => x.key))).toEqual([['bio'], ['city']]);
	});

	it('reports saving then saved, and has nothing unsaved afterwards', async () => {
		const { auto, statuses } = setup();
		auto.set(a('bio', 'x'));
		expect(auto.hasUnsaved).toBe(true);
		await vi.advanceTimersByTimeAsync(800);
		expect(statuses).toEqual(['saving', 'saved']);
		expect(auto.hasUnsaved).toBe(false);
	});

	it('never overlaps saves, and an edit made mid-save goes out next, not lost', async () => {
		let release!: () => void;
		const gate = new Promise<void>((r) => (release = r));
		const save = vi.fn().mockImplementationOnce(() => gate).mockResolvedValue(undefined);
		const { auto } = setup(save);

		auto.set(a('bio', 'first'));
		await vi.advanceTimersByTimeAsync(800); // first save starts and hangs
		auto.set(a('bio', 'second'));
		await vi.advanceTimersByTimeAsync(800); // timer fires while the first save is still running
		expect(save).toHaveBeenCalledTimes(1);

		release();
		await vi.advanceTimersByTimeAsync(0);
		expect(save).toHaveBeenCalledTimes(2);
		expect(save.mock.calls[1][0]).toEqual([a('bio', 'second')]);
		expect(auto.hasUnsaved).toBe(false);
	});

	it('keeps failed answers and resends them on the next flush', async () => {
		const save = vi.fn().mockRejectedValueOnce(new Error('offline')).mockResolvedValue(undefined);
		const { auto, statuses } = setup(save);
		auto.set(a('bio', 'precious'));
		await vi.advanceTimersByTimeAsync(800);
		expect(statuses).toEqual(['saving', 'error']);
		expect(auto.hasUnsaved).toBe(true);

		expect(await auto.flush()).toBe(true);
		expect(save).toHaveBeenCalledTimes(2);
		expect(save.mock.calls[1][0]).toEqual([a('bio', 'precious')]);
		expect(auto.hasUnsaved).toBe(false);
	});

	it('lets a newer edit win over a failed older one for the same key', async () => {
		let fail!: (e: Error) => void;
		const save = vi
			.fn()
			.mockImplementationOnce(() => new Promise<void>((_, rej) => (fail = rej)))
			.mockResolvedValue(undefined);
		const { auto } = setup(save);
		auto.set(a('bio', 'old'));
		await vi.advanceTimersByTimeAsync(800);
		auto.set(a('bio', 'newer')); // typed while the first save is in flight
		fail(new Error('boom'));
		await vi.advanceTimersByTimeAsync(0);
		expect(await auto.flush()).toBe(true);
		expect(save.mock.calls.at(-1)![0]).toEqual([a('bio', 'newer')]);
	});

	it('flush resolves false while saving keeps failing, and true with nothing pending', async () => {
		const { auto: idle } = setup();
		expect(await idle.flush()).toBe(true);

		const { auto } = setup(vi.fn().mockRejectedValue(new Error('down')));
		auto.set(a('bio', 'x'));
		expect(await auto.flush()).toBe(false);
		expect(auto.hasUnsaved).toBe(true);
	});

	it('flush saves immediately without waiting for the quiet period', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'x'));
		expect(await auto.flush()).toBe(true);
		expect(save).toHaveBeenCalledTimes(1);
	});

	it('does nothing after dispose', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'x'));
		auto.dispose();
		await vi.advanceTimersByTimeAsync(2000);
		auto.set(a('bio', 'y'));
		await vi.advanceTimersByTimeAsync(2000);
		expect(save).not.toHaveBeenCalled();
	});

	it('passes keepalive through when flushing because the page is going away', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'last words'));
		await auto.flush({ keepalive: true });
		expect(save).toHaveBeenCalledTimes(1);
		expect(save.mock.calls[0][1]).toEqual({ keepalive: true });
	});

	it('does not use keepalive for ordinary debounced saves', async () => {
		const { auto, save } = setup();
		auto.set(a('bio', 'x'));
		await vi.advanceTimersByTimeAsync(800);
		expect(save.mock.calls[0][1]).toEqual({});
	});
});
