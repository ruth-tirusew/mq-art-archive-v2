export type WikiSubmissionStatus = 'pending' | 'approved' | 'rejected';
export type WikiSubmissionKind = 'new' | 'edit';

export interface WikiSubmission {
  id: string;
  submitter_id: string;
  article_id?: string;
  title: string;
  body: string;
  kind: WikiSubmissionKind;
  based_on_version?: number;
  status: WikiSubmissionStatus;
  review_notes?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
}

// WikiSubmissionReview pairs a submission with the current state of its target article
// (undefined for a new-article submission), for rendering a diff.
export interface WikiSubmissionReview {
  submission: WikiSubmission;
  article?: { id: string; title: string; body: string; version: number };
}
