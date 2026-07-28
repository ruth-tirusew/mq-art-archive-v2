import { ProfileApi } from '$lib/adapters/api/profileApi';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { UpdateMyProfileInput } from '$lib/core/ports/profile';

const api = new ProfileApi();

export const profileService = {
  getArtistBySlug(slug: string): Promise<ArtistProfile> {
    return api.getArtistBySlug(slug);
  },
  getMyProfile(): Promise<ArtistProfile> {
    return api.getMyProfile();
  },
  updateMyProfile(input: UpdateMyProfileInput): Promise<ArtistProfile> {
    return api.updateMyProfile(input);
  }
};
