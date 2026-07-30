export type ApplicantType = 'artist' | 'institution';
export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

export interface OnboardingApplication {
	id: string;
	applicant_id?: string;
	applicant_type: ApplicantType;
	display_name: string;
	requested_handle?: string;
	notes?: string;
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
