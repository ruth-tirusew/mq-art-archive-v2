import { apiFetch } from '$lib/adapters/api/client';
import type { ArtistProfile, ArtistWrite, ProfileStatus } from '$lib/core/domain/artist';

export class ArtistsApi {
  list(status?: ProfileStatus): Promise<ArtistProfile[]> {
    const q = status ? `?status=${status}` : '';
    return apiFetch<ArtistProfile[]>(`/admin/v1/artists${q}`);
  }

  getById(id: string): Promise<ArtistProfile> {
    return apiFetch<ArtistProfile>(`/admin/v1/artists/${id}`);
  }

  create(body: ArtistWrite): Promise<ArtistProfile> {
    return apiFetch<ArtistProfile>(`/admin/v1/artists`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  update(id: string, body: ArtistWrite): Promise<ArtistProfile> {
    return apiFetch<ArtistProfile>(`/admin/v1/artists/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  patch(id: string, body: { status?: ProfileStatus; featured?: boolean }): Promise<ArtistProfile> {
    return apiFetch<ArtistProfile>(`/admin/v1/artists/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  delete(id: string): Promise<void> {
    return apiFetch<void>(`/admin/v1/artists/${id}`, { method: 'DELETE' });
  }
}
