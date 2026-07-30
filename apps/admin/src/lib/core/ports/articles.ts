import type {
  Article,
  ArticlePatch,
  ArticleRevision,
  ArticleStatus,
  ArticleWrite
} from '$lib/core/domain/article';

export interface ArticlesPort {
  list(status?: ArticleStatus): Promise<Article[]>;
  getById(id: string): Promise<Article>;
  create(write: ArticleWrite): Promise<Article>;
  update(id: string, write: ArticleWrite): Promise<Article>;
  patch(id: string, body: ArticlePatch): Promise<Article>;
  listRevisions(articleId: string): Promise<ArticleRevision[]>;
  getRevision(articleId: string, version: number): Promise<ArticleRevision>;
  restoreRevision(articleId: string, version: number): Promise<Article>;
}
