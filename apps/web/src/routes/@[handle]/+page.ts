import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { questionnaireService } from '$lib/composition';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => {
		const handle = params.handle;
		const profile = await artistsService.getByHandle(handle);
		const [posts, answers] = await Promise.all([
			artPostService.listByArtistSlug(profile.slug),
			questionnaireService.listArtistAnswers(profile.slug)
		]);
		return { source: 'api' as const, handle, artist: profile, posts, answers };
	}, 'We could not find that profile.');
