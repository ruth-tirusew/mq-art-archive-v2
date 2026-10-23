import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { profileService } from '$lib/application/profiles';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => {
		const [profile, posts, artists] = await Promise.all([
			profileService.getArtistBySlug(params.slug),
			artPostService.listByArtistSlug(params.slug),
			artistsService.list({ limit: 12 }).catch(() => [])
		]);

		const others = artists.filter((artist) => artist.slug !== params.slug).slice(0, 3);

		return { source: 'api' as const, artist: profile, posts, others };
	}, 'We could not find that artist.');
