export type ApplicantType = 'artist' | 'institution';
export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

/** An artist needs at least this many pieces of work to apply. Mirrors the API's rule. */
export const MIN_APPLICATION_WORKS = 5;
export const MAX_APPLICATION_WORKS = 20;

export interface ApplicantWork {
	id: string;
	media_asset_id: string;
	url: string;
	blurhash?: string;
	title: string;
	medium?: string;
	year?: number | null;
	position: number;
}

export interface AddWorkInput {
	media_asset_id: string;
	title: string;
	medium?: string;
	year?: number | null;
}

/**
 * A piece of work the applicant has chosen but not uploaded. Nothing leaves the browser until they
 * submit: the image stays a local file with a local preview, and is uploaded (and attached to the
 * application) only as part of submitting.
 */
export interface WorkDraft {
	key: string;
	file: File;
	/** An object URL for previewing the local file. Revoke it when the draft goes away. */
	previewUrl: string;
	title: string;
	medium: string;
	/** As typed; validated by workDraftProblem. */
	year: string;
	status: 'pending' | 'uploading' | 'done' | 'failed';
	/** The stored work's id once uploaded, so it can still be removed. */
	workId?: string;
	error?: string;
}

/** What is wrong with a draft's text fields, or null. The image itself is checked on selection. */
export function workDraftProblem(draft: Pick<WorkDraft, 'title' | 'year'>, now = new Date()): string | null {
	if (!draft.title.trim()) return 'Give this piece a title.';
	if (draft.title.trim().length > 200) return 'Keep the title under 200 characters.';
	const year = draft.year.trim();
	if (year) {
		const n = Number(year);
		if (!Number.isInteger(n) || n < 1000 || n > now.getFullYear() + 1) return 'Year should look like 2024.';
	}
	return null;
}

/** How many more works are needed before the applicant can submit, or 0. */

export function worksShortfall(have: number): number {
	return Math.max(0, MIN_APPLICATION_WORKS - have);
}

export interface OnboardingApplication {
	id: string;
	applicant_id?: string;
	applicant_type: ApplicantType;
	display_name: string;
	requested_handle?: string;
	notes?: string;
	review_notes?: string;
	status: ApprovalStatus;
	reviewed_by?: string | null;
	reviewed_at?: string | null;
	/** Set when the application was submitted with questionnaire answers. */
	questionnaire_revision_id?: string | null;
	/** The work submitted with the application, as it was at submission. */
	works?: ApplicantWork[];
	created_at?: string;
	updated_at?: string;
}

export interface SubmitApplicationInput {
	applicant_type: ApplicantType;
	display_name: string;
	requested_handle?: string;
	notes?: string;
}

export interface HandleAvailability {
	handle: string;
	available: boolean;
}

// Mirrors the API's rules (backend onboarding usecase). The server stays authoritative: these only
// let the form say so before a round trip.
export const MIN_PASSWORD_LENGTH = 8;

const HANDLE_PATTERN = /^[a-z0-9_]{3,30}$/;

/** Handles are case-insensitive; the API stores them lowercase. */
export function normalizeHandle(raw: string): string {
	return raw.trim().toLowerCase();
}

export function isValidHandle(handle: string): boolean {
	return HANDLE_PATTERN.test(handle);
}

/** The plain (no questionnaire) application folds its fields into one free-text note. */
export function composeLinkNotes(fields: {
	instagram: string;
	website: string;
	telegram: string;
	about: string;
}): string {
	const { instagram, website, telegram, about } = fields;
	return [
		instagram.trim() && `Instagram: ${instagram.trim()}`,
		website.trim() && `Website: ${website.trim()}`,
		telegram.trim() && `Telegram: ${telegram.trim()}`,
		about.trim() && `About my practice:\n${about.trim()}`
	]
		.filter(Boolean)
		.join('\n');
}

/**
 * Inverse of composeLinkNotes, so a rejected applicant's earlier note refills the form's fields
 * instead of being pasted whole into "about" (which would nest the labels on the next submit).
 * Text that isn't in the composed shape is kept intact as "about".
 */
export function parseLinkNotes(notes: string): {
	instagram: string;
	website: string;
	telegram: string;
	about: string;
} {
	const fields = { instagram: '', website: '', telegram: '', about: '' };
	const aboutMarker = 'About my practice:\n';
	const marked = notes.startsWith(aboutMarker) ? 0 : notes.indexOf('\n' + aboutMarker) + 1 || -1;
	const hasAbout = marked >= 0;
	const head = hasAbout ? notes.slice(0, marked) : notes;
	const labels = { 'Instagram: ': 'instagram', 'Website: ': 'website', 'Telegram: ': 'telegram' } as const;

	const leftovers: string[] = [];
	for (const line of head.split('\n')) {
		const label = (Object.keys(labels) as (keyof typeof labels)[]).find((l) => line.startsWith(l));
		if (label && !fields[labels[label]]) fields[labels[label]] = line.slice(label.length);
		else if (line.trim()) leftovers.push(line);
	}
	if (hasAbout) {
		fields.about = notes.slice(marked + aboutMarker.length);
		// anything unrecognised in front of the marker is not lost
		if (leftovers.length) fields.about = `${leftovers.join('\n')}\n${fields.about}`;
	} else {
		fields.about = leftovers.join('\n');
	}
	return fields;
}
