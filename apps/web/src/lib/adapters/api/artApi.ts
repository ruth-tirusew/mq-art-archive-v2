import { apiFetch } from '$lib/adapters/api/client';
import { normalizeArtPost, normalizeArtPosts } from '$lib/adapters/api/normalize';
import type { ArtListParams, ArtPost } from '$lib/core/domain/art';
import type { ArtPort } from '$lib/core/ports/art';

function toQuery(params?: ArtListParams): string {
  if (!params) return '';
  const q = new URLSearchParams();
  if (params.city) q.set('city', params.city);
  if (params.medium) q.set('medium', params.medium);
  if (params.year) q.set('year', String(params.year));
  if (params.style) q.set('style', params.style);
  if (params.featured !== undefined) q.set('featured', String(params.featured));
  if (params.q) q.set('q', params.q);
  if (params.limit) q.set('limit', String(params.limit));
  if (params.offset) q.set('offset', String(params.offset));
  const s = q.toString();
  return s ? `?${s}` : '';
}

export class ArtApi implements ArtPort {
  async list(params?: ArtListParams): Promise<ArtPost[]> {
    const raw = await apiFetch<unknown>(`/api/v1/posts${toQuery(params)}`);
    return normalizeArtPosts(raw);
  }

  async listByArtistSlug(slug: string): Promise<ArtPost[]> {
    const raw = await apiFetch<unknown>(`/api/v1/artists/${slug}/posts`);
    return normalizeArtPosts(raw);
  }

  async getById(id: string): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/posts/${id}`);
    return normalizeArtPost(raw);
  }
}
