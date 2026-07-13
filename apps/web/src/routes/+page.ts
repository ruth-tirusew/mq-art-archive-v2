import { artPostService } from '$lib/application/artPosts';
import { articleService } from '$lib/application/articles';
import { artistsService } from '$lib/application/artists';
import { eventsService } from '$lib/application/events';
import { marqueeItemsFromApiEvents } from '$lib/utils/marquee';
import type { PageLoad } from './$types';

function pickFeatured<T extends { slug: string; featured?: boolean }>(artists: T[]) {
	const featured = artists.find((artist) => artist.featured) ?? null;
	const roster = (featured ? artists.filter((artist) => artist.slug !== featured.slug) : artists).slice(
		0,
		4
	);
	return { featured, roster };
}

export const load: PageLoad = async () => {
	const [artists, acquisitions, articles, upcomingEvents] = await Promise.all([
		artistsService.list({ limit: 20 }).catch(() => []),
		artPostService
			.list({ limit: 12, featured: true })
			.catch(() => artPostService.list({ limit: 12 }).catch(() => [])),
		articleService.listPublished({ limit: 3 }).catch(() => []),
		eventsService.list({ upcoming: true, limit: 20 }).catch(() => [])
	]);

	const { featured: featuredArtist, roster } = pickFeatured(artists);
	const featuredPosts = featuredArtist
		? await artPostService.listByArtistSlug(featuredArtist.slug).catch(() => [])
		: [];

	return {
		source: 'api' as const,
		featuredArtist,
		featuredPosts,
		acquisitions,
		artists: roster,
		articles,
		marqueeItems: marqueeItemsFromApiEvents(upcomingEvents),
		featured: Boolean(featuredArtist)
	};
};
