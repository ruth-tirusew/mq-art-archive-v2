import type { AnswerInput } from '$lib/core/domain/questionnaire';

export type SaveStatus = 'idle' | 'saving' | 'saved' | 'error';

export interface SaveContext {
	/**
	 * True when the page is being hidden or closed. The request must be sent with `keepalive` so it
	 * outlives the page (browsers cap such bodies at 64KB; a failure here is just a lost last edit).
	 */
	keepalive?: boolean;
}

export interface AutosaveOptions {
	/** Persist these answers together (all or none). Reject to signal failure. */
	save: (answers: AnswerInput[], context: SaveContext) => Promise<void>;
	/** Quiet period after the last edit before saving. */
	delayMs?: number;
	onStatus?: (status: SaveStatus, error?: string) => void;
}

/**
 * Debounced, ordered autosave for a form with many independent answers.
 *
 *  - Edits to the same key coalesce; only what changed is sent.
 *  - Saves never overlap: edits made during a save are sent in the next round, not lost.
 *  - A failed save keeps its answers pending, so the next edit (or retry/flush) sends them again,
 *    unless the artist has typed something newer for that key, which then wins.
 *  - flush() resolves true only when nothing is left unsaved.
 */
export class AnswerAutosave {
	private pending = new Map<string, AnswerInput>();
	private timer: ReturnType<typeof setTimeout> | undefined;
	private running: Promise<boolean> | null = null;
	private disposed = false;

	constructor(private readonly opts: AutosaveOptions) {}

	/** True while anything is waiting to be saved or being saved. */
	get hasUnsaved(): boolean {
		return this.pending.size > 0 || this.running !== null;
	}

	/** Queue the latest value for a key and (re)start the quiet period. */
	set(answer: AnswerInput): void {
		if (this.disposed) return;
		this.pending.set(answer.key, answer);
		clearTimeout(this.timer);
		this.timer = setTimeout(() => void this.run({}), this.opts.delayMs ?? 800);
	}

	/** Stop waiting: save everything now. Resolves true if nothing remains unsaved. */
	async flush(context: SaveContext = {}): Promise<boolean> {
		clearTimeout(this.timer);
		for (;;) {
			const ok = await this.run(context);
			if (!ok) return false;
			if (this.pending.size === 0) return true;
		}
	}

	/** Drop timers; nothing more will be sent. Unsaved answers are discarded. */
	dispose(): void {
		this.disposed = true;
		clearTimeout(this.timer);
		this.pending.clear();
	}

	private run(context: SaveContext): Promise<boolean> {
		if (this.running) return this.running;
		if (this.pending.size === 0) return Promise.resolve(true);

		this.running = (async () => {
			while (this.pending.size > 0 && !this.disposed) {
				const sent = new Map(this.pending);
				this.pending.clear();
				this.opts.onStatus?.('saving');
				try {
					await this.opts.save([...sent.values()], context);
					this.opts.onStatus?.('saved');
				} catch (err) {
					// Put back what failed, unless a newer edit for the same key arrived meanwhile.
					for (const [key, answer] of sent) if (!this.pending.has(key)) this.pending.set(key, answer);
					this.opts.onStatus?.('error', err instanceof Error ? err.message : 'Could not save');
					return false;
				}
			}
			return true;
		})().finally(() => {
			this.running = null;
		});
		return this.running;
	}
}
