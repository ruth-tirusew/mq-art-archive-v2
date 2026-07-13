import { EventsApi } from '$lib/adapters/api/eventsApi';
import type { Event, EventListParams } from '$lib/core/domain/events';

const api = new EventsApi();

export const eventsService = {
  list(params?: EventListParams): Promise<Event[]> {
    return api.list(params);
  },
  getBySlug(slug: string): Promise<Event> {
    return api.getBySlug(slug);
  }
};
