import type { ArtPost } from '$lib/core/domain/art';
import type { Article } from '$lib/core/domain/content';
import type { ArtistProfile } from '$lib/core/domain/profile';
import { postImageUrl } from '$lib/utils/display';

/** Typed cells in a curated exhibition spread — matches the mockup art direction. */
export type CuratorModuleKind =
	| 'curators-note'
	| 'collection-highlight'
	| 'artist-quote'
	| 'exhibition'
	| 'featured-essay'
	| 'from-archive';

export type CuratorModule = {
	kind: CuratorModuleKind;
	eyebrow: string;
	title: string;
	body: string;
	href: string;
	cta: string;
};

export type ArtistTile = {
	post: ArtPost;
	badge?: string;
};

/**
 * One homepage spread — magazine issue, not a masonry shuffle.
 * Desktop geometry lives in CSS via `layoutId`; this holds content only.
 */
export type EditorialSpread = {
	layoutId: string;
	intro: {
		eyebrow: string;
		headline: string;
		body: string;
		ctaLabel: string;
		ctaHref: string;
	};
	feature: ArtistTile;
	support: ArtistTile;
	archive: ArtistTile;
	emerging: ArtistTile;
	note: CuratorModule;
};

const INTRO = {
	eyebrow: "Today's curation",
	headline: 'Discover Ethiopian artists. Explore their worlds.',
	body: 'A living archive of modern Ethiopian art and its diaspora — portraits, practices, and the conversations that hold them.',
	ctaLabel: 'Explore all artists',
	ctaHref: '/artists'
} as const;

const CURATOR_FALLBACKS: CuratorModule[] = [
	{
		kind: 'curators-note',
		eyebrow: "Curator's note",
		title: 'Fragmented histories, living practices.',
		body: 'How contemporary Ethiopian artists reclaim lineage without treating the past as a museum case.',
		href: '/wiki',
		cta: 'Read note'
	},
	{
		kind: 'collection-highlight',
		eyebrow: 'Collection highlight',
		title: 'Works that refuse a single reading.',
		body: 'A short tour through recent acquisitions that sit between painting, ritual, and public memory.',
		href: '/archive',
		cta: 'See highlight'
	},
	{
		kind: 'artist-quote',
		eyebrow: 'Quote from an artist',
		title: '"The archive is not behind us — it walks with us."',
		body: 'Voices from the studio on inheritance, material, and what it means to be seen from Addis and abroad.',
		href: '/artists',
		cta: 'Meet artists'
	},
	{
		kind: 'exhibition',
		eyebrow: 'Exhibition announcement',
		title: "What's opening in Addis this month.",
		body: 'Openings, walkthroughs, and residencies — verified with institutional partners.',
		href: '/events',
		cta: 'See calendar'
	},
	{
		kind: 'featured-essay',
		eyebrow: 'Featured essay',
		title: 'On pigment, paperwork, and presence.',
		body: 'Essays from the community handbook — contracts, materials, and the work of being an artist in Ethiopia.',
		href: '/wiki',
		cta: 'Read essay'
	},
	{
		kind: 'from-archive',
		eyebrow: 'From the archive',
		title: 'A quieter corner of the collection.',
		body: 'Older works pulled forward — not as nostalgia, but as context for what is being made now.',
		href: '/archive',
		cta: 'Browse archive'
	}
];

/** Named spreads — Editorial A / B / C (and more). Same components, different hangings. */
export const EDITORIAL_LAYOUT_IDS = [
	'editorial-a', // Text | Hero | Artist / Archive | Note | Emerging  (mockup)
	'editorial-b', // Full-bleed hero / Artist | Note | Emerging
	'editorial-c' // Hero | Artist / three-up bottom
] as const;

export type EditorialLayoutId = (typeof EDITORIAL_LAYOUT_IDS)[number];

function hasImage(post: ArtPost): boolean {
	return Boolean(postImageUrl(post.media, post.id, post.artist_slug));
}

function artistKey(post: ArtPost): string {
	return post.artist_slug || post.artist_id || post.id;
}

function shuffle<T>(items: T[], seed: number): T[] {
	const out = [...items];
	let s = seed % 2147483647;
	if (s <= 0) s += 2147483646;
	for (let i = out.length - 1; i > 0; i--) {
		s = (s * 16807) % 2147483647;
		const j = s % (i + 1);
		[out[i], out[j]] = [out[j]!, out[i]!];
	}
	return out;
}

function uniqueByArtist(posts: ArtPost[]): ArtPost[] {
	const seen = new Set<string>();
	const out: ArtPost[] = [];
	for (const post of posts) {
		if (!hasImage(post)) continue;
		const key = artistKey(post);
		if (seen.has(key)) continue;
		seen.add(key);
		out.push(post);
	}
	return out;
}

function moduleFromArticle(article: Article): CuratorModule {
	return {
		kind: 'featured-essay',
		eyebrow: article.category?.trim() || 'Featured essay',
		title: article.title,
		body: article.excerpt?.trim() || article.body.slice(0, 140).trim(),
		href: `/wiki/${article.slug}`,
		cta: 'Read note'
	};
}

function pickNote(articles: Article[], seed: number, index: number): CuratorModule {
	if (articles.length > 0) {
		const article = articles[(seed + index) % articles.length]!;
		return moduleFromArticle(article);
	}
	return CURATOR_FALLBACKS[(seed + index) % CURATOR_FALLBACKS.length]!;
}

/**
 * Build curated homepage spreads from the artist pool + optional essays.
 * Layout geometry rotates; content roles stay stable.
 */
export function buildEditorialSpreads(
	posts: ArtPost[],
	opts?: {
		spreadCount?: number;
		seed?: number;
		featuredArtist?: ArtistProfile | null;
		featuredPosts?: ArtPost[];
		articles?: Article[];
	}
): EditorialSpread[] {
	const pool = uniqueByArtist([
		...(opts?.featuredPosts ?? []),
		...posts.filter((p) => p.featured_acquisition),
		...posts
	]);
	if (pool.length === 0) return [];

	const seed =
		typeof opts?.seed === 'number' && Number.isFinite(opts.seed)
			? Math.abs(Math.floor(opts.seed))
			: Math.floor(Math.random() * 1_000_000);

	const spreadCount = Math.min(opts?.spreadCount ?? 4, 8);
	const articles = opts?.articles ?? [];
	const featuredSlug = opts?.featuredArtist?.slug;
	const featuredWorks = (opts?.featuredPosts ?? []).filter(hasImage);

	const spreads: EditorialSpread[] = [];

	for (let i = 0; i < spreadCount; i++) {
		const rotated = shuffle(pool, seed + i * 97);
		// First spread always matches the mockup (Editorial A); later ones rotate.
		const layoutId =
			i === 0
				? 'editorial-a'
				: EDITORIAL_LAYOUT_IDS[(seed + i) % EDITORIAL_LAYOUT_IDS.length]!;

		const featurePost =
			(featuredWorks.length > 0 ? featuredWorks[i % featuredWorks.length] : undefined) ??
			(featuredSlug ? rotated.find((p) => p.artist_slug === featuredSlug) : undefined) ??
			rotated[0]!;

		const rest = rotated.filter((p) => p.id !== featurePost.id);
		while (rest.length < 3) {
			rest.push(featurePost);
		}

		const [supportPost, archivePost, emergingPost] = rest;

		spreads.push({
			layoutId,
			intro: { ...INTRO },
			feature: { post: featurePost, badge: 'Featured' },
			support: { post: supportPost!, badge: undefined },
			archive: { post: archivePost!, badge: 'From the archive' },
			emerging: { post: emergingPost!, badge: 'Emerging voice' },
			note: pickNote(articles, seed, i)
		});
	}

	return spreads;
}

/** @deprecated Use EditorialSpread — kept for any leftover imports during transition. */
export type EditorialWall = EditorialSpread;
export type EditorialLayout = { id: string };
export const EDITORIAL_LAYOUTS = EDITORIAL_LAYOUT_IDS.map((id) => ({ id }));

export function layoutById(id: string | undefined): { id: string } {
	return { id: id && EDITORIAL_LAYOUT_IDS.includes(id as EditorialLayoutId) ? id : 'editorial-a' };
}

export function buildEditorialWalls(
	posts: ArtPost[],
	opts?: Parameters<typeof buildEditorialSpreads>[1]
): EditorialSpread[] {
	return buildEditorialSpreads(posts, opts);
}
