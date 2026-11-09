import type { WorkDraft } from '$lib/core/domain/onboarding';

/**
 * Keeps an applicant's chosen-but-not-uploaded works in this browser (IndexedDB), so a refresh or
 * a closed tab doesn't throw away their images and titles. Nothing goes to the server: uploading
 * still happens only on submit, on purpose, so swapped images never cost storage.
 *
 * Every call is best-effort. Where IndexedDB is missing or refuses (private mode, a full disk),
 * load returns nothing and save reports false, and the apply form works exactly as without it.
 */

/** What is stored for one draft. A File survives IndexedDB's structured clone as a Blob. */
export interface StoredDraft {
	key: string;
	file: Blob;
	name: string;
	type: string;
	title: string;
	medium: string;
	year: string;
	/** Set once the piece reached the server (a submit that failed part way). */
	workId?: string;
}

interface StoredRecord {
	userId: string;
	savedAt: number;
	drafts: StoredDraft[];
}

const DB = 'artiv-apply';
const STORE = 'work-drafts';
/** Drafts nobody has touched for this long are dropped rather than restored. */
const MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

export function toStored(drafts: WorkDraft[]): StoredDraft[] {
	return drafts.map((d) => ({
		key: d.key,
		file: d.file,
		name: d.file.name,
		type: d.file.type,
		title: d.title,
		medium: d.medium,
		year: d.year,
		...(d.workId ? { workId: d.workId } : {})
	}));
}

/**
 * Rebuilds drafts with fresh preview URLs. A piece that was mid-upload when the page died is
 * pending again, unless it had already been stored, in which case it is done and won't upload twice.
 */
export function fromStored(stored: StoredDraft[], previewUrl: (file: File) => string = (f) => URL.createObjectURL(f)): WorkDraft[] {
	return stored.map((s) => {
		const file = s.file instanceof File ? s.file : new File([s.file], s.name, { type: s.type || s.file.type });
		return {
			key: s.key,
			file,
			previewUrl: previewUrl(file),
			title: s.title,
			medium: s.medium,
			year: s.year,
			status: s.workId ? 'done' : 'pending',
			...(s.workId ? { workId: s.workId } : {})
		};
	});
}

export function isFresh(savedAt: number, now = Date.now()): boolean {
	return now - savedAt <= MAX_AGE_MS;
}

function open(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		if (typeof indexedDB === 'undefined') return reject(new Error('IndexedDB unavailable'));
		const req = indexedDB.open(DB, 1);
		req.onupgradeneeded = () => req.result.createObjectStore(STORE, { keyPath: 'userId' });
		req.onsuccess = () => resolve(req.result);
		req.onerror = () => reject(req.error);
	});
}

async function run<T>(mode: IDBTransactionMode, op: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
	const db = await open();
	try {
		return await new Promise<T>((resolve, reject) => {
			const tx = db.transaction(STORE, mode);
			const req = op(tx.objectStore(STORE));
			tx.oncomplete = () => resolve(req.result);
			tx.onerror = () => reject(tx.error);
			tx.onabort = () => reject(tx.error);
		});
	} finally {
		db.close();
	}
}

/** The portrait draft is kept under its own key so it never disturbs the works record. */
interface StoredPortrait {
	userId: string;
	savedAt: number;
	portrait: { file: Blob; name: string; type: string };
}

const portraitKey = (userId: string) => `${userId}#portrait`;

export const workDraftStore = {
	async loadPortrait(userId: string): Promise<File | null> {
		try {
			const record = await run<StoredPortrait | undefined>('readonly', (s) => s.get(portraitKey(userId)));
			if (!record) return null;
			if (!isFresh(record.savedAt)) {
				await this.savePortrait(userId, null);
				return null;
			}
			const { file, name, type } = record.portrait;
			return file instanceof File ? file : new File([file], name, { type: type || file.type });
		} catch {
			return null;
		}
	},

	/** True when the portrait is now safe on this device. Null clears it. */
	async savePortrait(userId: string, file: File | null): Promise<boolean> {
		try {
			if (!file) {
				await run('readwrite', (s) => s.delete(portraitKey(userId)));
				return true;
			}
			const record: StoredPortrait = {
				userId: portraitKey(userId),
				savedAt: Date.now(),
				portrait: { file, name: file.name, type: file.type }
			};
			await run('readwrite', (s) => s.put(record));
			return true;
		} catch {
			return false;
		}
	},

	async load(userId: string): Promise<WorkDraft[]> {
		try {
			const record = await run<StoredRecord | undefined>('readonly', (s) => s.get(userId));
			if (!record) return [];
			if (!isFresh(record.savedAt)) {
				await this.clear(userId);
				return [];
			}
			return fromStored(record.drafts);
		} catch {
			return [];
		}
	},

	/** True when the drafts are now safe on this device. An empty list clears the record. */
	async save(userId: string, drafts: WorkDraft[]): Promise<boolean> {
		try {
			if (drafts.length === 0) {
				await this.clear(userId);
				return true;
			}
			const record: StoredRecord = { userId, savedAt: Date.now(), drafts: toStored(drafts) };
			await run('readwrite', (s) => s.put(record));
			return true;
		} catch {
			return false;
		}
	},

	async clear(userId: string): Promise<void> {
		try {
			await run('readwrite', (s) => s.delete(userId));
		} catch {
			// Nothing to clean up if storage never worked.
		}
	}
};
