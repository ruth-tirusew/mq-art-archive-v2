import { PostsApi } from '$lib/adapters/api/postsApi';
import type { ArtPost, ArtPostCreate, ArtPostWrite, ArtStatus } from '$lib/core/domain/art';

const api = new PostsApi();

export const postsService = {
  list(status?: ArtStatus): Promise<ArtPost[]> {
    return api.list(status);
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
