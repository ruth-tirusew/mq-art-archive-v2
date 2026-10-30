import { PostsApi } from '$lib/adapters/api/postsApi';
import { fetchAll } from '$lib/adapters/api/pageQuery';
import type { PageParams } from '$lib/adapters/api/pageQuery';
import type { ArtPost, ArtPostCreate, ArtPostWrite, ArtStatus } from '$lib/core/domain/art';

const api = new PostsApi();

export const postsService = {
  list(status?: ArtStatus, page?: PageParams): Promise<ArtPost[]> {
    return api.list(status, page);
  },
  /** Every row, fetched in pages. For dashboards that count rather than browse. */
  listAll(status?: ArtStatus): Promise<ArtPost[]> {
    return fetchAll((page) => api.list(status, page));
  },
  getById(id: string): Promise<ArtPost> {
    return api.getById(id);
  },
  create(body: ArtPostCreate): Promise<ArtPost> {
    return api.create(body);
  },
  update(id: string, body: ArtPostWrite): Promise<ArtPost> {
    return api.update(id, body);
  },
  patch(
    id: string,
    body: { status?: ArtStatus; featured_acquisition?: boolean }
  ): Promise<ArtPost> {
    return api.patch(id, body);
  },
  delete(id: string): Promise<void> {
    return api.delete(id);
  }
};
