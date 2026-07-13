import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import type { ArtPost } from '$lib/core/domain/art';
import type { PageLoad } from './$types';

function groupApiPosts(posts: ArtPost[]): Record<string, ArtPost[]> {
	const map: Record<string, ArtPost[]> = {};
	for (const post of posts) {
		const slug = post.artist_slug;
		if (!slug) continue;
		(map[slug] ??= []).push(post);
	}
	return map;
}

export const load: PageLoad = async () => {
	const [artists, posts] = await Promise.all([
		artistsService.list({ limit: 100 }).catch(() => []),
		artPostService.list({ limit: 200 }).catch(() => [])
	]);
	const roster = artists.filter((a) => a.handle !== 'demo' && a.slug !== 'demo-yordanos-kebede');
	const rosterPosts = posts.filter((p) => p.artist_slug !== 'demo-yordanos-kebede');

	return {
		source: 'api' as const,
		artists: roster,
		postsBySlug: groupApiPosts(rosterPosts),
		totalWorks: rosterPosts.length
	};
};
