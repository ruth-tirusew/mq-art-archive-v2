import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import type { PageLoad } from './$types';

const DEMO_HANDLE = 'demo';

export const load: PageLoad = async () => {
	const profile = await artistsService.getByHandle(DEMO_HANDLE);
	const posts = await artPostService.listByArtistSlug(profile.slug).catch(() => []);

	return {
		handle: DEMO_HANDLE,
		artist: profile,
		posts
	};
};
