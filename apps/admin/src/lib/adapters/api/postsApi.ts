import { apiFetch } from '$lib/adapters/api/client';
import { pageQuery, type PageParams } from '$lib/adapters/api/pageQuery';
import type { ArtPost, ArtPostCreate, ArtPostWrite, ArtStatus } from '$lib/core/domain/art';

export class PostsApi {
  list(status?: ArtStatus, page?: PageParams): Promise<ArtPost[]> {
    return apiFetch<ArtPost[]>(`/admin/v1/posts${pageQuery(status, page)}`);
  }

  getById(id: string): Promise<ArtPost> {
    return apiFetch<ArtPost>(`/admin/v1/posts/${id}`);
  }

  create(body: ArtPostCreate): Promise<ArtPost> {
    return apiFetch<ArtPost>(`/admin/v1/posts`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  update(id: string, body: ArtPostWrite): Promise<ArtPost> {
    return apiFetch<ArtPost>(`/admin/v1/posts/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  patch(
    id: string,
    body: { status?: ArtStatus; featured_acquisition?: boolean }
  ): Promise<ArtPost> {
    return apiFetch<ArtPost>(`/admin/v1/posts/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  delete(id: string): Promise<void> {
    return apiFetch<void>(`/admin/v1/posts/${id}`, { method: 'DELETE' });
  }
}
