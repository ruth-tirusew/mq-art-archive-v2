import type { Event, EventWrite, ReviewPayload, ReviewStatus, SyncEventsResult } from '$lib/core/domain/event';

export interface EventsPort {
  listPending(): Promise<Event[]>;
  getById(id: string): Promise<Event>;
  review(id: string, status: ReviewStatus, notes: string): Promise<Event>;
  sync(): Promise<SyncEventsResult>;
}

export type { Event, EventWrite, ReviewPayload, ReviewStatus, SyncEventsResult };
