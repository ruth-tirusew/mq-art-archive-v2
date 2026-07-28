import type { ArtListParams, ArtPost } from '$lib/core/domain/art';

export interface CreateArtDraftInput {
  title: string;
  description?: string;
  medium?: string;
  year?: number | null;
  dimensions?: string;
  city?: string;
  style?: string;
  palette?: string[];
  media_urls?: string[];
}

export type UpdateArtPostInput = CreateArtDraftInput;

export interface ArtPort {
  list(params?: ArtListParams): Promise<ArtPost[]>;
  listByArtistSlug(slug: string): Promise<ArtPost[]>;
  getById(id: string): Promise<ArtPost>;
  listMyPosts(): Promise<ArtPost[]>;
  getMyPost(id: string): Promise<ArtPost>;
  createDraft(input: CreateArtDraftInput): Promise<ArtPost>;
  updateMyPost(id: string, input: UpdateArtPostInput): Promise<ArtPost>;
  publishMyPost(id: string): Promise<ArtPost>;
  unpublishMyPost(id: string): Promise<ArtPost>;
  archiveMyPost(id: string): Promise<ArtPost>;
  deleteMyPost(id: string): Promise<void>;
}
