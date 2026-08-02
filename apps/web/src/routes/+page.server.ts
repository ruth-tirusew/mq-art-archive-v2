import { artPostService } from '$lib/application/artPosts';
import { articleService } from '$lib/application/articles';
import { artistsService } from '$lib/application/artists';
import { eventsService } from '$lib/application/events';
import { buildEditorialSpreads } from '$lib/components/home/editorialCompositions';
import { getApiBaseUrl } from '$lib/config/dataSource';
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

const emptyPosts: ArtPost[] = [];

async function sleep(ms: number) {
	await new Promise((resolve) => setTimeout(resolve, ms));
}

/** Render cold starts can fail the first parallel batch; retry post reads briefly. */
async function loadPosts<T>(load: () => Promise<T>, attempts = 3): Promise<T> {
	let last: unknown;
	for (let i = 0; i < attempts; i++) {
		try {
			return await load();
		} catch (err) {
			last = err;
			if (i < attempts - 1) await sleep(400 * (i + 1));
		}
	}
	throw last;
}

export const load: PageServerLoad = async ({ fetch }) => {
	const empty = [] as const;

	// Wake the API before fan-out (helps Render free tier cold start on Vercel SSR).
	await fetch(`${getApiBaseUrl()}/health`, { signal: AbortSignal.timeout(25_000) }).catch(() => {});

	const [artistsResult, acquisitionsResult, articlesResult, upcomingEventsResult] =
		await Promise.allSettled([
			artistsService.list({ limit: 20 }),
			loadPosts(() => artPostService.list({ limit: 12, featured: true })),
			articleService.listPublished({ limit: 6 }),
			eventsService.list({ upcoming: true, limit: 20 })
		]);

	const artists =
		artistsResult.status === 'fulfilled' ? artistsResult.value : [...empty];
	let acquisitions =
		acquisitionsResult.status === 'fulfilled' ? acquisitionsResult.value : [...emptyPosts];

	if (acquisitions.length === 0) {
		acquisitions = await loadPosts(() => artPostService.list({ limit: 12 })).catch(
			() => [...emptyPosts]
		);
	}

	const articles =
		articlesResult.status === 'fulfilled' ? articlesResult.value : [...empty];
	const upcomingEvents =
		upcomingEventsResult.status === 'fulfilled' ? upcomingEventsResult.value : [...empty];

	const featuredArtist = pickFeatured(artists);

	const [featuredPosts, coverPosts] = await Promise.all([
		featuredArtist
			? loadPosts(() => artPostService.listByArtistSlug(featuredArtist.slug)).catch(
					() => [...emptyPosts]
				)
			: Promise.resolve([...emptyPosts]),
		loadPosts(() => artPostService.list({ limit: 80 })).catch(() => [...acquisitions])
	]);

	const canvasPool = mergePosts(acquisitions, coverPosts, featuredPosts);
	let editorialSpreads = buildEditorialSpreads(canvasPool, {
		spreadCount: 4,
		featuredArtist,
		featuredPosts,
		articles
	});

	if (editorialSpreads.length === 0) {
		const rescue = await loadPosts(() => artPostService.list({ limit: 80 })).catch(
			() => [...emptyPosts]
		);
		if (rescue.length > 0) {
			editorialSpreads = buildEditorialSpreads(mergePosts(rescue, featuredPosts), {
				spreadCount: 4,
				featuredArtist,
				featuredPosts,
				articles
			});
		}
	}

	return {
		source: 'api' as const,
		featuredArtist,
		featuredPosts,
		acquisitions,
		editorialSpreads,
		/** @deprecated — prefer editorialSpreads */
		editorialWalls: editorialSpreads,
		/** Set on the server so SSR does not rely on reactive spread length (Svelte 5 + SK). */
		showEditorialHero: editorialSpreads.length > 0,
		articles,
		marqueeItems: marqueeItemsFromApiEvents(upcomingEvents),
		featured: Boolean(featuredArtist)
	};
};
