import { ArtApi } from '$lib/adapters/api/artApi';
import type { ArtListParams, ArtPost } from '$lib/core/domain/art';

const api = new ArtApi();

export const artPostService = {
  list(params?: ArtListParams): Promise<ArtPost[]> {
    return api.list(params);
  },
  listByArtistSlug(slug: string): Promise<ArtPost[]> {
    return api.listByArtistSlug(slug);
  },
  getById(id: string): Promise<ArtPost> {
    return api.getById(id);
  }
};
