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
  /** The questionnaire revision the applicant answered, if any. */
  questionnaire_revision_id?: string;
  /** Set when the applicant arrived through an invite link. */
  invite_id?: string;
  created_at?: string;
  updated_at?: string;
}

export const REJECTION_NOTE_REQUIRED = 'Add a reason before rejecting. The applicant will see it.';

/** The reason a decision can't be sent yet, or null when it can. */
export function reviewBlocker(status: Exclude<ApprovalStatus, 'pending'>, notes: string): string | null {
  return status === 'rejected' && !notes.trim() ? REJECTION_NOTE_REQUIRED : null;
}
