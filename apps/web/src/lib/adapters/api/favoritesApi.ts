import { apiFetch } from '$lib/adapters/api/client';
import type { Article } from '$lib/core/domain/content';
import type { FavoriteStatus } from '$lib/core/domain/favorite';

export class FavoritesApi {
	getStatus(articleId: string): Promise<FavoriteStatus> {
		return apiFetch<FavoriteStatus>(`/api/v1/favorites/${articleId}`);
	}

	toggle(articleId: string): Promise<FavoriteStatus> {
		return apiFetch<FavoriteStatus>(`/api/v1/me/favorites/${articleId}`, { method: 'POST' });
	}

	listMine(): Promise<Article[]> {
		return apiFetch<Article[]>('/api/v1/me/favorites');
	}
}
