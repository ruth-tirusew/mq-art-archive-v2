import { apiFetch } from '$lib/adapters/api/client';
import { normalizeArtPost, normalizeArtPosts } from '$lib/adapters/api/normalize';
import type { ArtListParams, ArtPost } from '$lib/core/domain/art';
import type { ArtPort, CreateArtDraftInput, UpdateArtPostInput } from '$lib/core/ports/art';

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

function toBody(input: CreateArtDraftInput | UpdateArtPostInput) {
  return {
    title: input.title,
    description: input.description ?? '',
    medium: input.medium ?? '',
    year: input.year ?? null,
    dimensions: input.dimensions ?? '',
    city: input.city ?? '',
    style: input.style ?? '',
    palette: input.palette ?? [],
    media_urls: input.media_urls ?? []
  };
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

  async listMyPosts(): Promise<ArtPost[]> {
    const raw = await apiFetch<unknown>('/api/v1/me/posts');
    return normalizeArtPosts(raw);
  }

  async getMyPost(id: string): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/me/posts/${id}`);
    return normalizeArtPost(raw);
  }

  async createDraft(input: CreateArtDraftInput): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>('/api/v1/posts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(toBody(input))
    });
    return normalizeArtPost(raw);
  }

  async updateMyPost(id: string, input: UpdateArtPostInput): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/me/posts/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(toBody(input))
    });
    return normalizeArtPost(raw);
  }

  async publishMyPost(id: string): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/me/posts/${id}/publish`, {
      method: 'POST'
    });
    return normalizeArtPost(raw);
  }

  async unpublishMyPost(id: string): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/me/posts/${id}/unpublish`, {
      method: 'POST'
    });
    return normalizeArtPost(raw);
  }

  async archiveMyPost(id: string): Promise<ArtPost> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/me/posts/${id}/archive`, {
      method: 'POST'
    });
    return normalizeArtPost(raw);
  }

  async deleteMyPost(id: string): Promise<void> {
    await apiFetch<unknown>(`/api/v1/me/posts/${id}`, { method: 'DELETE' });
  }
}
