import { WikiSubmissionsApi } from '$lib/adapters/api/wikiSubmissionsApi';
import type { SubmitWikiInput, WikiSubmission } from '$lib/core/domain/wikiSubmission';

const api = new WikiSubmissionsApi();

export const wikiSubmissionsService = {
	listMine(): Promise<WikiSubmission[]> {
		return api.listMine();
	},
	submit(input: SubmitWikiInput): Promise<WikiSubmission> {
		return api.submit(input);
	}
};
