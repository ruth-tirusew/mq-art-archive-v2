import { apiFetch } from '$lib/adapters/api/client';
import { normalizeArticle, normalizeArticles } from '$lib/adapters/api/normalize';
import type { Article, ArticleListParams } from '$lib/core/domain/content';
import type { ContentPort } from '$lib/core/ports/content';

function toQuery(params?: ArticleListParams): string {
  if (!params) return '';
  const q = new URLSearchParams();
  if (params.category) q.set('category', params.category);
  if (params.q) q.set('q', params.q);
  if (params.limit) q.set('limit', String(params.limit));
  if (params.offset) q.set('offset', String(params.offset));
  const s = q.toString();
  return s ? `?${s}` : '';
}

export class ArticleApi implements ContentPort {
  listPublished(params?: ArticleListParams): Promise<Article[]> {
    return apiFetch<unknown>(`/api/v1/articles${toQuery(params)}`).then(normalizeArticles);
  }

  getBySlug(slug: string): Promise<Article> {
    return apiFetch<Record<string, unknown>>(`/api/v1/articles/${slug}`).then(normalizeArticle);
  }

  createDraft(title: string, body: string): Promise<Article> {
    return apiFetch<Record<string, unknown>>('/api/v1/articles', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title, body })
    }).then(normalizeArticle);
  }
}
