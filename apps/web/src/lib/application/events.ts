import { EventsApi } from '$lib/adapters/api/eventsApi';
import type { Event, EventListParams, EventSubmission } from '$lib/core/domain/events';

const api = new EventsApi();

export const eventsService = {
  list(params?: EventListParams): Promise<Event[]> {
    return api.list(params);
  },
  getBySlug(slug: string): Promise<Event> {
    return api.getBySlug(slug);
  },
  submit(payload: EventSubmission): Promise<Event> {
    return api.submit(payload);
  }
};
