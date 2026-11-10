import { apiFetch } from '$lib/adapters/api/client';
import { pageQuery, type PageParams } from '$lib/adapters/api/pageQuery';
import type { Event, EventStatus, EventWrite, ReviewStatus, SyncEventsResult } from '$lib/core/domain/event';

export class EventsApi {
  list(status?: EventStatus | 'all', page?: PageParams): Promise<Event[]> {
    return apiFetch<Event[]>(`/admin/v1/events${pageQuery(status, page)}`);
  }

  getById(id: string): Promise<Event> {
    return apiFetch<Event>(`/admin/v1/events/${id}`);
  }

  create(body: EventWrite): Promise<Event> {
    return apiFetch<Event>(`/admin/v1/events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  update(id: string, body: EventWrite): Promise<Event> {
    return apiFetch<Event>(`/admin/v1/events/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  review(id: string, status: ReviewStatus, notes: string): Promise<Event> {
    return apiFetch<Event>(`/admin/v1/events/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status, notes })
    });
  }

  delete(id: string): Promise<void> {
    return apiFetch<void>(`/admin/v1/events/${id}`, { method: 'DELETE' });
  }

  sync(): Promise<SyncEventsResult> {
    return apiFetch<SyncEventsResult>('/admin/v1/events/sync', { method: 'POST' });
  }
}
