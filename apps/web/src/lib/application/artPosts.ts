import { ArtApi } from '$lib/adapters/api/artApi';
import type { ArtListParams, ArtPost } from '$lib/core/domain/art';
import type { CreateArtDraftInput, UpdateArtPostInput } from '$lib/core/ports/art';

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
  },
  listMyPosts(): Promise<ArtPost[]> {
    return api.listMyPosts();
  },
  getMyPost(id: string): Promise<ArtPost> {
    return api.getMyPost(id);
  },
  createDraft(input: CreateArtDraftInput): Promise<ArtPost> {
    return api.createDraft(input);
  },
  updateMyPost(id: string, input: UpdateArtPostInput): Promise<ArtPost> {
    return api.updateMyPost(id, input);
  },
  publishMyPost(id: string): Promise<ArtPost> {
    return api.publishMyPost(id);
  },
  unpublishMyPost(id: string): Promise<ArtPost> {
    return api.unpublishMyPost(id);
  },
  archiveMyPost(id: string): Promise<ArtPost> {
    return api.archiveMyPost(id);
  },
  deleteMyPost(id: string): Promise<void> {
    return api.deleteMyPost(id);
  }
};
