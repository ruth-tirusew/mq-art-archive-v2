import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => {
		const handle = params.handle;
		const profile = await artistsService.getByHandle(handle);
		const posts = await artPostService.listByArtistSlug(profile.slug);
		return { source: 'api' as const, handle, artist: profile, posts };
	}, 'We could not find that profile.');
