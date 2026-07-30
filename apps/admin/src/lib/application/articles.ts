import { ArticlesApi } from '$lib/adapters/api/articlesApi';
import type {
  Article,
  ArticlePatch,
  ArticleRevision,
  ArticleStatus,
  ArticleWrite
} from '$lib/core/domain/article';

const api = new ArticlesApi();

export const articlesService = {
  list(status?: ArticleStatus): Promise<Article[]> {
    return api.list(status);
  },
  getById(id: string): Promise<Article> {
    return api.getById(id);
  },
  create(write: ArticleWrite): Promise<Article> {
    return api.create(write);
  },
  update(id: string, write: ArticleWrite): Promise<Article> {
    return api.update(id, write);
  },
  patch(id: string, body: ArticlePatch): Promise<Article> {
    return api.patch(id, body);
  },
  listRevisions(articleId: string): Promise<ArticleRevision[]> {
    return api.listRevisions(articleId);
  },
  getRevision(articleId: string, version: number): Promise<ArticleRevision> {
    return api.getRevision(articleId, version);
  },
  restoreRevision(articleId: string, version: number): Promise<Article> {
    return api.restoreRevision(articleId, version);
  }
};
