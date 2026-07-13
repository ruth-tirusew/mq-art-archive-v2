import type { Article, ArticleListParams } from '$lib/core/domain/content';

export interface ContentPort {
  listPublished(params?: ArticleListParams): Promise<Article[]>;
  getBySlug(slug: string): Promise<Article>;
  createDraft(title: string, body: string): Promise<Article>;
}
