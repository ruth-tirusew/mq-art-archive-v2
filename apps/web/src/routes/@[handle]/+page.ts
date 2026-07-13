import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const handle = params.handle;

	try {
		const profile = await artistsService.getByHandle(handle);
		const posts = await artPostService.listByArtistSlug(profile.slug);
		return { source: 'api' as const, handle, artist: profile, posts };
	} catch {
		error(404, 'Profile not found');
	}
};
