import { apiFetch } from '$lib/adapters/api/client';
import type { WikiSubmission } from '$lib/core/domain/wikiSubmission';

export class WikiSubmissionsApi {
  listPending(): Promise<WikiSubmission[]> {
    return apiFetch<WikiSubmission[]>('/admin/v1/wiki/submissions');
  }

  review(id: string, action: 'approve' | 'reject', notes = ''): Promise<WikiSubmission> {
    return apiFetch<WikiSubmission>(`/admin/v1/wiki/submissions/${id}/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ notes })
    });
  }
}
