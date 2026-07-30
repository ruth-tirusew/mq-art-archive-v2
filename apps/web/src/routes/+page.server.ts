import { artPostService } from '$lib/application/artPosts';
import { articleService } from '$lib/application/articles';
import { artistsService } from '$lib/application/artists';
import { eventsService } from '$lib/application/events';
import { buildEditorialSpreads } from '$lib/components/home/editorialCompositions';
import { marqueeItemsFromApiEvents } from '$lib/utils/marquee';
import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { PageServerLoad } from './$types';

function pickFeatured(artists: ArtistProfile[]) {
	return artists.find((artist) => artist.featured) ?? null;
}

/** Dedupe posts while preferring earlier (featured) entries. */
function mergePosts(...groups: ArtPost[][]): ArtPost[] {
	const seen = new Set<string>();
	const out: ArtPost[] = [];
	for (const group of groups) {
		for (const post of group) {
			if (seen.has(post.id)) continue;
			seen.add(post.id);
			out.push(post);
		}
	}
	return out;
}

export const load: PageServerLoad = async () => {
	const empty = [] as const;

	const [artistsResult, acquisitionsResult, articlesResult, upcomingEventsResult] =
		await Promise.allSettled([
			artistsService.list({ limit: 20 }),
			artPostService.list({ limit: 12, featured: true }),
			articleService.listPublished({ limit: 6 }),
			eventsService.list({ upcoming: true, limit: 20 })
		]);

	const artists =
		artistsResult.status === 'fulfilled' ? artistsResult.value : [...empty];
	let acquisitions =
		acquisitionsResult.status === 'fulfilled' ? acquisitionsResult.value : [...empty];

	if (acquisitions.length === 0) {
		acquisitions = await artPostService.list({ limit: 12 }).catch(() => [...empty]);
	}

	const articles =
		articlesResult.status === 'fulfilled' ? articlesResult.value : [...empty];
	const upcomingEvents =
		upcomingEventsResult.status === 'fulfilled' ? upcomingEventsResult.value : [...empty];

	const featuredArtist = pickFeatured(artists);

	const [featuredPosts, coverPosts] = await Promise.all([
		featuredArtist
			? artPostService.listByArtistSlug(featuredArtist.slug).catch(() => [...empty])
			: Promise.resolve([...empty]),
		artPostService.list({ limit: 80 }).catch(() => acquisitions)
	]);

	const canvasPool = mergePosts(acquisitions, coverPosts);
	const editorialSpreads = buildEditorialSpreads(canvasPool, {
		spreadCount: 4,
		featuredArtist,
		featuredPosts,
		articles
	});

	return {
		source: 'api' as const,
		featuredArtist,
		featuredPosts,
		acquisitions,
		editorialSpreads,
		/** @deprecated — prefer editorialSpreads */
		editorialWalls: editorialSpreads,
		articles,
		marqueeItems: marqueeItemsFromApiEvents(upcomingEvents),
		featured: Boolean(featuredArtist)
	};
};
