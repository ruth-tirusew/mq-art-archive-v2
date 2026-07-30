import { apiFetch } from '$lib/adapters/api/client';
import type { SubmitWikiInput, WikiSubmission } from '$lib/core/domain/wikiSubmission';

export class WikiSubmissionsApi {
	listMine(): Promise<WikiSubmission[]> {
		return apiFetch<WikiSubmission[]>('/api/v1/me/wiki/submissions');
	}

	submit(input: SubmitWikiInput): Promise<WikiSubmission> {
		return apiFetch<WikiSubmission>('/api/v1/me/wiki/submissions', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		});
	}
}
