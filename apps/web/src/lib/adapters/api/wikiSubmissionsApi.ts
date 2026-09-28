import { apiFetch } from '$lib/adapters/api/client';
import type {
	SubmitWikiInput,
	WikiSubmission,
	WikiSubmissionReview
} from '$lib/core/domain/wikiSubmission';

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

	// listOwnPending, getForReview, approve and reject are the artist-facing review actions,
	// scoped by the backend to submissions against an article the caller owns.
	listOwnPending(): Promise<WikiSubmission[]> {
		return apiFetch<WikiSubmission[]>('/api/v1/me/wiki/submissions/pending');
	}

	getForReview(id: string): Promise<WikiSubmissionReview> {
		return apiFetch<WikiSubmissionReview>(`/api/v1/me/wiki/submissions/${id}`);
	}

	review(id: string, action: 'approve' | 'reject', notes = ''): Promise<WikiSubmission> {
		return apiFetch<WikiSubmission>(`/api/v1/me/wiki/submissions/${id}/${action}`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ notes })
		});
	}
}
