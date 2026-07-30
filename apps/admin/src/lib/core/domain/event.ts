export type EventStatus = 'pending' | 'approved' | 'rejected';

export interface Event {
  id: string;
  slug: string;
  title: string;
  description?: string;
  source_url?: string;
  image_url?: string | null;
  event_type: string;
  venue?: string;
  city?: string;
  starts_at: string;
  ends_at?: string | null;
  status: EventStatus;
  review_notes?: string;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  scraped_at?: string | null;
  created_at?: string;
  updated_at?: string;
}

export type ReviewStatus = 'approved' | 'rejected';

export interface ReviewPayload {
  status: ReviewStatus;
  notes?: string;
}

export interface EventWrite {
  title: string;
  description?: string;
  source_url?: string;
  image_url?: string | null;
  event_type?: string;
  venue?: string;
  city?: string;
  slug?: string;
  starts_at: string;
  ends_at?: string | null;
  status?: EventStatus;
}

export interface SyncEventsResult {
  upserted: number;
}
