import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import type { ArtistListParams, ArtistProfile } from '$lib/core/domain/profile';

const api = new ArtistsApi();

export const artistsService = {
  list(params?: ArtistListParams): Promise<ArtistProfile[]> {
    return api.list(params);
  },
  getByHandle(handle: string): Promise<ArtistProfile> {
    return api.getByHandle(handle);
  }
};
