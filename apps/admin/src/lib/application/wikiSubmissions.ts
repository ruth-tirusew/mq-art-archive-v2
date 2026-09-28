import { WikiSubmissionsApi } from '$lib/adapters/api/wikiSubmissionsApi';

const api = new WikiSubmissionsApi();

export const wikiSubmissionsService = {
  listPending: () => api.listPending(),
  getForReview: (id: string) => api.getForReview(id),
  approve: (id: string, notes?: string) => api.review(id, 'approve', notes),
  reject: (id: string, notes?: string) => api.review(id, 'reject', notes)
};
