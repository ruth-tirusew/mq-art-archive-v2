import { apiFetch } from '$lib/adapters/api/client';
import type { Highlight, PopularHighlight } from '$lib/core/domain/highlight';

export class HighlightsApi {
	create(articleId: string, quotedText: string, start: number, end: number): Promise<Highlight> {
		return apiFetch<Highlight>(`/api/v1/wiki/articles/${articleId}/highlights`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ quoted_text: quotedText, start, end })
		});
	}

	delete(highlightId: string): Promise<void> {
		return apiFetch<void>(`/api/v1/wiki/highlights/${highlightId}`, { method: 'DELETE' });
	}

	listMine(articleId: string): Promise<Highlight[]> {
		return apiFetch<Highlight[]>(`/api/v1/wiki/articles/${articleId}/highlights/mine`);
	}

	getPopular(articleId: string): Promise<PopularHighlight | null> {
		return apiFetch<PopularHighlight | null>(`/api/v1/wiki/articles/${articleId}/highlights/popular`);
	}
}
