import { ArticleApi } from '$lib/adapters/api/articleApi';
import type { Article, ArticleListParams } from '$lib/core/domain/content';

const api = new ArticleApi();

export const articleService = {
  listPublished(params?: ArticleListParams): Promise<Article[]> {
    return api.listPublished(params);
  },
  getBySlug(slug: string): Promise<Article> {
    return api.getBySlug(slug);
  }
};
