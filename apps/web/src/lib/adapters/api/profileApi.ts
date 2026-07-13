import { apiFetch } from '$lib/adapters/api/client';
import { normalizeProfile } from '$lib/adapters/api/normalize';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { ProfilePort } from '$lib/core/ports/profile';

export class ProfileApi implements ProfilePort {
  async getArtistBySlug(slug: string): Promise<ArtistProfile> {
    const raw = await apiFetch<Record<string, unknown>>(`/api/v1/artists/${slug}`);
    return normalizeProfile(raw);
  }
}
