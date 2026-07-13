import type { ArtistProfile } from '$lib/core/domain/profile';

export interface ProfilePort {
  getArtistBySlug(slug: string): Promise<ArtistProfile>;
}
