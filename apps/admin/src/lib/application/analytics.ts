import { apiFetch } from '$lib/adapters/api/client';

export interface AnalyticsSummary {
  total: number;
  artists: number;
  posts: number;
  articles: number;
}

type AnalyticsRow = { entity_type?: string; count?: number };

export async function getAnalyticsSummary(): Promise<AnalyticsSummary> {
  const response = await apiFetch<AnalyticsRow[] | Record<string, number>>('/admin/v1/analytics');
  if (Array.isArray(response)) {
    const summary = { total: 0, artists: 0, posts: 0, articles: 0 };
    for (const row of response) {
      const count = Number(row.count) || 0;
      summary.total += count;
      if (row.entity_type === 'artist') summary.artists += count;
      if (row.entity_type === 'post') summary.posts += count;
      if (row.entity_type === 'article') summary.articles += count;
    }
    return summary;
  }
  return {
    total: Number(response.total ?? response.views) || 0,
    artists: Number(response.artists ?? response.artist) || 0,
    posts: Number(response.posts ?? response.post) || 0,
    articles: Number(response.articles ?? response.article) || 0
  };
}
