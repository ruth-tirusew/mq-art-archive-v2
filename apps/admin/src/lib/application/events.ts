import { EventsApi } from '$lib/adapters/api/eventsApi';
import type { EventStatus, EventWrite, ReviewStatus } from '$lib/core/domain/event';

const api = new EventsApi();

export const eventsService = {
  list(status?: EventStatus | 'all') {
    return api.list(status);
  },
  listPending() {
    return api.list('pending');
  },
  getById(id: string) {
    return api.getById(id);
  },
  create(body: EventWrite) {
    return api.create(body);
  },
  update(id: string, body: EventWrite) {
    return api.update(id, body);
  },
  review(id: string, status: ReviewStatus, notes: string) {
    return api.review(id, status, notes);
  },
  delete(id: string) {
    return api.delete(id);
  },
  sync() {
    return api.sync();
  }
};
