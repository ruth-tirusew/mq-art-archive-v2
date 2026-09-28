import { WikiSubmissionsApi } from '$lib/adapters/api/wikiSubmissionsApi';
import type {
	SubmitWikiInput,
	WikiSubmission,
	WikiSubmissionReview
} from '$lib/core/domain/wikiSubmission';

const api = new WikiSubmissionsApi();

export const wikiSubmissionsService = {
	listMine(): Promise<WikiSubmission[]> {
		return api.listMine();
	},
	submit(input: SubmitWikiInput): Promise<WikiSubmission> {
		return api.submit(input);
	},
	listOwnPending(): Promise<WikiSubmission[]> {
		return api.listOwnPending();
	},
	getForReview(id: string): Promise<WikiSubmissionReview> {
		return api.getForReview(id);
	},
	approve(id: string, notes?: string): Promise<WikiSubmission> {
		return api.review(id, 'approve', notes);
	},
	reject(id: string, notes?: string): Promise<WikiSubmission> {
		return api.review(id, 'reject', notes);
	}
};
