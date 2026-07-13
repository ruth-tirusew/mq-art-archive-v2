export type EventStatus = 'pending' | 'approved' | 'rejected';

export interface Event {
  id: string;
  slug: string;
  title: string;
  description?: string;
  image_url?: string | null;
  event_type: string;
  venue?: string;
  city?: string;
  starts_at: string;
  ends_at?: string | null;
  status: EventStatus;
  created_at?: string;
  updated_at?: string;
}

export interface EventListParams {
  type?: string;
  q?: string;
  upcoming?: boolean;
  limit?: number;
  offset?: number;
}
