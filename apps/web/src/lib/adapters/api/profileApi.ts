import { apiFetch } from '$lib/adapters/api/client';
import { normalizeProfile } from '$lib/adapters/api/normalize';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { ProfilePort, UpdateMyProfileInput } from '$lib/core/ports/profile';

export class ProfileApi implements ProfilePort {
  async getArtistBySlug(slug: string): Promise<ArtistProfile> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/artists/${slug}`);
    return normalizeProfile(raw);
  }

  async getMyProfile(): Promise<ArtistProfile> {
    const raw = await apiFetch<Record<string, unknown>>('/api/v1/me/profile');
    return normalizeProfile(raw);
  }

  async updateMyProfile(input: UpdateMyProfileInput): Promise<ArtistProfile> {
    const raw = await apiFetch<Record<string, unknown>>('/api/v1/me/profile', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input)
    });
    return normalizeProfile(raw);
  }
}
