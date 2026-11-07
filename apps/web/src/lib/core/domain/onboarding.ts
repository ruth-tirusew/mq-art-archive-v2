export type ApplicantType = 'artist' | 'institution';
export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

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
