import { WikiSubmissionsApi } from '$lib/adapters/api/wikiSubmissionsApi';

const api = new WikiSubmissionsApi();

export const wikiSubmissionsService = {
  listPending: () => api.listPending(),
  approve: (id: string, notes?: string) => api.review(id, 'approve', notes),
  reject: (id: string, notes?: string) => api.review(id, 'reject', notes)
};
