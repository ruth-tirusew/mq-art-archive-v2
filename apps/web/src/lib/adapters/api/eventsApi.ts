import { apiFetch } from '$lib/adapters/api/client';
import { normalizeEvent, normalizeEvents } from '$lib/adapters/api/normalize';
import type { Event, EventListParams } from '$lib/core/domain/events';

function toQuery(params?: EventListParams): string {
  if (!params) return '';
  const q = new URLSearchParams();
  if (params.type) q.set('type', params.type);
  if (params.q) q.set('q', params.q);
  if (params.upcoming === false) q.set('upcoming', 'false');
  if (params.limit) q.set('limit', String(params.limit));
  if (params.offset) q.set('offset', String(params.offset));
  const s = q.toString();
  return s ? `?${s}` : '';
}

export class EventsApi {
  list(params?: EventListParams): Promise<Event[]> {
    return apiFetch<unknown>(`/api/v1/events${toQuery(params)}`).then(normalizeEvents);
  }

  getBySlug(slug: string): Promise<Event> {
    return apiFetch<Record<string, unknown>>(`/api/v1/events/${slug}`).then(normalizeEvent);
  }
}
