import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { fetchAll } from '$lib/adapters/api/pageQuery';
import type { PageParams } from '$lib/adapters/api/pageQuery';
import type { ArtistProfile, ArtistWrite, ProfileStatus } from '$lib/core/domain/artist';

const api = new ArtistsApi();

export const artistsService = {
  list(status?: ProfileStatus, page?: PageParams): Promise<ArtistProfile[]> {
    return api.list(status, page);
  },
  /** Every row, fetched in pages. For dashboards that count rather than browse. */
  listAll(status?: ProfileStatus): Promise<ArtistProfile[]> {
    return fetchAll((page) => api.list(status, page));
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
