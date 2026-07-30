import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import type { ArtistProfile, ArtistWrite, ProfileStatus } from '$lib/core/domain/artist';

const api = new ArtistsApi();

export const artistsService = {
  list(status?: ProfileStatus): Promise<ArtistProfile[]> {
    return api.list(status);
  },
  getById(id: string): Promise<ArtistProfile> {
    return api.getById(id);
  },
  create(body: ArtistWrite): Promise<ArtistProfile> {
    return api.create(body);
  },
  update(id: string, body: ArtistWrite): Promise<ArtistProfile> {
    return api.update(id, body);
  },
  patch(id: string, body: { status?: ProfileStatus; featured?: boolean }): Promise<ArtistProfile> {
    return api.patch(id, body);
  },
  delete(id: string): Promise<void> {
    return api.delete(id);
  }
};
