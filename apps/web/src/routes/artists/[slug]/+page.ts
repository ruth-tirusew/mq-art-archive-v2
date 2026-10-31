import { artPostService } from '$lib/application/artPosts';
import { artistsService } from '$lib/application/artists';
import { profileService } from '$lib/application/profiles';
import { questionnaireService } from '$lib/application/questionnaire';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => {
		const [profile, posts, artists, answers] = await Promise.all([
			profileService.getArtistBySlug(params.slug),
			artPostService.listByArtistSlug(params.slug),
			artistsService.list({ limit: 12 }).catch(() => []),
			questionnaireService.listArtistAnswers(params.slug)
		]);

		const others = artists.filter((artist) => artist.slug !== params.slug).slice(0, 3);

		return { source: 'api' as const, artist: profile, posts, others, answers };
	}, 'We could not find that artist.');
