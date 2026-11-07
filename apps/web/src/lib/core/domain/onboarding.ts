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
