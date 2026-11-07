export type ApplicantType = 'artist' | 'institution';
export type ApprovalStatus = 'pending' | 'approved' | 'rejected';

/** What the admin list can show: one status, or every application. */
export type ApplicationStatusFilter = ApprovalStatus | 'all';

export const APPLICATION_STATUS_FILTERS: readonly ApplicationStatusFilter[] = ['pending', 'approved', 'rejected', 'all'];

export function parseApplicationStatusFilter(raw: string | null): ApplicationStatusFilter {
  return APPLICATION_STATUS_FILTERS.find((s) => s === raw) ?? 'pending';
}

export interface ApplicationPage {
  items: OnboardingApplication[];
  /** Matches across all pages, not just this one. */
  total: number;
}

export const APPLICATIONS_PAGE_SIZE = 50;

/** An artist needs at least this many pieces of work to apply. Mirrors the API's rule. */
export const MIN_APPLICATION_WORKS = 5;

export interface ApplicantWork {
  id: string;
  url: string;
  blurhash?: string;
  title: string;
  medium?: string;
  year?: number | null;
  position: number;
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
  /** The questionnaire revision the applicant answered, if any. */
  questionnaire_revision_id?: string;
  /** Set when the applicant arrived through an invite link. */
  invite_id?: string;
  /** The work submitted with the application, as it was at submission. */
  works?: ApplicantWork[];
  created_at?: string;
  updated_at?: string;
}

export const REJECTION_NOTE_REQUIRED = 'Add a reason before rejecting. The applicant will see it.';

/** The reason a decision can't be sent yet, or null when it can. */
export function reviewBlocker(status: Exclude<ApprovalStatus, 'pending'>, notes: string): string | null {
  return status === 'rejected' && !notes.trim() ? REJECTION_NOTE_REQUIRED : null;
}
