import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { ArtApi } from '$lib/adapters/api/artApi';
import { ArticleApi } from '$lib/adapters/api/articleApi';
import { EventsApi } from '$lib/adapters/api/eventsApi';

const artistsApi = new ArtistsApi();
const artApi = new ArtApi();
const articleApi = new ArticleApi();
const eventsApi = new EventsApi();

export type SearchResults = {
  artists: Awaited<ReturnType<ArtistsApi['list']>>;
  posts: Awaited<ReturnType<ArtApi['list']>>;
  articles: Awaited<ReturnType<ArticleApi['listPublished']>>;
  events: Awaited<ReturnType<EventsApi['list']>>;
};

export async function searchAll(query: string, limit = 5): Promise<SearchResults> {
  const q = query.trim();
  if (!q) {
    return { artists: [], posts: [], articles: [], events: [] };
  }

  const [artists, posts, articles, events] = await Promise.all([
    artistsApi.list({ q, limit }).catch(() => []),
    artApi.list({ q, limit }).catch(() => []),
    articleApi.listPublished({ q, limit }).catch(() => []),
    eventsApi.list({ q, limit, upcoming: false }).catch(() => [])
  ]);

  return { artists, posts, articles, events };
}
