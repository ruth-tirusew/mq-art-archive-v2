import { HighlightsApi } from '$lib/adapters/api/highlightsApi';
import type { ActivityHighlight } from '$lib/core/domain/activity';
import type { Highlight, PopularHighlight } from '$lib/core/domain/highlight';

const api = new HighlightsApi();

export const highlightsService = {
	create(articleId: string, quotedText: string, start: number, end: number): Promise<Highlight> {
		return api.create(articleId, quotedText, start, end);
	},
	delete(highlightId: string): Promise<void> {
		return api.delete(highlightId);
	},
	listMine(articleId: string): Promise<Highlight[]> {
		return api.listMine(articleId);
	},
	getPopular(articleId: string): Promise<PopularHighlight | null> {
		return api.getPopular(articleId);
	},
	listMyActivity(): Promise<ActivityHighlight[]> {
		return api.listMyActivity();
	}
};
