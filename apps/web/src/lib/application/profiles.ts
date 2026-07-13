import { ProfileApi } from '$lib/adapters/api/profileApi';
import type { ArtistProfile } from '$lib/core/domain/profile';

const api = new ProfileApi();

export const profileService = {
  getArtistBySlug(slug: string): Promise<ArtistProfile> {
    return api.getArtistBySlug(slug);
  }
};
