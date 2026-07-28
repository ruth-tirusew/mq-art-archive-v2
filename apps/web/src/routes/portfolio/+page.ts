import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const profiles = await artistsService.list({ featured: true, limit: 1 }).catch(() => []);
	const profile = profiles[0] ?? null;
	const posts = profile
		? await artPostService.listByArtistSlug(profile.slug).catch(() => [])
		: [];

	return {
		handle: profile?.handle ?? profile?.slug ?? null,
		artist: profile,
		posts
	};
};
