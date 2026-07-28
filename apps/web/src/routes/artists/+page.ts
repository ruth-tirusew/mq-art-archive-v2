import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import type { ArtPost } from '$lib/core/domain/art';
import type { PageLoad } from './$types';

export type ArtistsSortKey = 'name' | 'works' | 'recent';

function groupApiPosts(posts: ArtPost[]): Record<string, ArtPost[]> {
	const map: Record<string, ArtPost[]> = {};
	for (const post of posts) {
		const slug = post.artist_slug;
		if (!slug) continue;
		(map[slug] ??= []).push(post);
	}
	return map;
}

function parseSort(raw: string | null): ArtistsSortKey {
	if (raw === 'works' || raw === 'recent' || raw === 'name') return raw;
	return 'name';
}

export const load: PageLoad = async ({ url }) => {
	const page = Math.max(1, Number(url.searchParams.get('page') ?? 1) || 1);
	const limit = 24;
	const [artistsResult, postsResult] = await Promise.all([
		artistsService.listPage({ limit, offset: (page - 1) * limit }).then(
			(result) => ({ ok: true as const, artists: result.data, total: result.total }),
			() => ({ ok: false as const, artists: [], total: 0 })
		),
		artPostService.list({ limit: 200 }).then(
			(posts) => ({ ok: true as const, posts }),
			() => ({ ok: false as const, posts: [] })
		)
	]);

	const roster = artistsResult.artists;
	const rosterPosts = postsResult.posts;
	const loadFailed = !artistsResult.ok && !postsResult.ok;

	const discipline = url.searchParams.get('discipline') || null;
	const sort = parseSort(url.searchParams.get('sort'));

	return {
		source: loadFailed ? ('unavailable' as const) : ('api' as const),
		artists: roster,
		postsBySlug: groupApiPosts(rosterPosts),
		totalWorks: rosterPosts.length,
		loadFailed,
		pagination: { page, limit, total: artistsResult.total, pages: Math.max(1, Math.ceil(artistsResult.total / limit)) },
		filters: { discipline, sort }
	};
};
