export type WikiSubmissionStatus = 'pending' | 'approved' | 'rejected';

export interface WikiSubmission {
  id: string;
  submitter_id: string;
  article_id?: string;
  title: string;
  body: string;
  status: WikiSubmissionStatus;
  review_notes?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
  updated_at: string;
}
