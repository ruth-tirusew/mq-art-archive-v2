import type { ArtListParams, ArtPost } from '$lib/core/domain/art';

export interface ArtPort {
  list(params?: ArtListParams): Promise<ArtPost[]>;
  listByArtistSlug(slug: string): Promise<ArtPost[]>;
  getById(id: string): Promise<ArtPost>;
}
