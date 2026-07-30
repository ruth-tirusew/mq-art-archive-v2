import { apiFetch } from '$lib/adapters/api/client';
import type {
  Article,
  ArticlePatch,
  ArticleRevision,
  ArticleStatus,
  ArticleWrite
} from '$lib/core/domain/article';
import type { ArticlesPort } from '$lib/core/ports/articles';

export class ArticlesApi implements ArticlesPort {
  list(status?: ArticleStatus): Promise<Article[]> {
    const q = status ? `?status=${status}` : '';
    return apiFetch<Article[]>(`/admin/v1/articles${q}`);
  }

  getById(id: string): Promise<Article> {
    return apiFetch<Article>(`/admin/v1/articles/${id}`);
  }

  create(write: ArticleWrite): Promise<Article> {
    return apiFetch<Article>('/admin/v1/articles', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(write)
    });
  }

  update(id: string, write: ArticleWrite): Promise<Article> {
    return apiFetch<Article>(`/admin/v1/articles/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(write)
    });
  }

  patch(id: string, body: ArticlePatch): Promise<Article> {
    return apiFetch<Article>(`/admin/v1/articles/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  listRevisions(articleId: string): Promise<ArticleRevision[]> {
    return apiFetch<ArticleRevision[]>(`/admin/v1/articles/${articleId}/revisions`);
  }

  getRevision(articleId: string, version: number): Promise<ArticleRevision> {
    return apiFetch<ArticleRevision>(`/admin/v1/articles/${articleId}/revisions/${version}`);
  }

  restoreRevision(articleId: string, version: number): Promise<Article> {
    return apiFetch<Article>(`/admin/v1/articles/${articleId}/revisions/${version}/restore`, {
      method: 'POST'
    });
  }
}
