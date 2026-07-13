import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { profileService } from '$lib/application/profiles';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		const [profile, posts, artists] = await Promise.all([
			profileService.getArtistBySlug(params.slug),
			artPostService.listByArtistSlug(params.slug),
			artistsService.list({ limit: 12 }).catch(() => [])
		]);

		const others = artists.filter((artist) => artist.slug !== params.slug).slice(0, 3);

		return { source: 'api' as const, artist: profile, posts, others };
	} catch {
		error(404, 'Artist not found');
	}
};
