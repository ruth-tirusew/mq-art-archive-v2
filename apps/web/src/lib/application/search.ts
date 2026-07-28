import { apiFetch } from '$lib/adapters/api/client';
import {
  normalizeArticles,
  normalizeProfiles,
  normalizeArtPosts,
  normalizeEvents
} from '$lib/adapters/api/normalize';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { ArtPost } from '$lib/core/domain/art';
import type { Article } from '$lib/core/domain/content';
import type { Event } from '$lib/core/domain/events';

export type SearchResults = {
  artists: ArtistProfile[];
  posts: ArtPost[];
  articles: Article[];
  events: Event[];
};

export async function searchAll(query: string, limit = 5): Promise<SearchResults> {
  const q = query.trim();
  if (!q) {
    return { artists: [], posts: [], articles: [], events: [] };
  }

  const response = await apiFetch<Record<string, unknown>>(
    `/api/v1/search?q=${encodeURIComponent(q)}&limit=${limit}`
  );
  return {
    artists: normalizeProfiles(response.artists),
    posts: normalizeArtPosts(response.posts),
    articles: normalizeArticles(response.articles),
    events: normalizeEvents(response.events)
  };
}
