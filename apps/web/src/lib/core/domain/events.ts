export type EventStatus = 'pending' | 'approved' | 'rejected';

export interface Event {
  id: string;
  slug: string;
  title: string;
  description?: string;
  image_url?: string | null;
  source_url?: string | null;
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

export interface EventSubmission {
  title: string;
  description?: string;
  source_url?: string;
  event_type?: string;
  venue?: string;
  city?: string;
  starts_at: string;
  ends_at?: string | null;
}

export function sourceLabel(sourceUrl?: string | null): string {
  if (!sourceUrl) return 'Partner';
  const lower = sourceUrl.toLowerCase();
  if (lower.startsWith('admin://') || lower.startsWith('public://')) return 'Partner';
  if (lower.includes('t.me/') || lower.includes('telegram')) return 'Telegram Channel';
  if (lower.includes('facebook.com') || lower.includes('fb.com')) return 'Facebook Page';
  return 'Website';
}
