import { FavoritesApi } from '$lib/adapters/api/favoritesApi';
import type { Article } from '$lib/core/domain/content';
import type { FavoriteStatus } from '$lib/core/domain/favorite';

const api = new FavoritesApi();

export const favoritesService = {
	getStatus(articleId: string): Promise<FavoriteStatus> {
		return api.getStatus(articleId);
	},
	toggle(articleId: string): Promise<FavoriteStatus> {
		return api.toggle(articleId);
	},
	listMine(): Promise<Article[]> {
		return api.listMine();
	}
};
