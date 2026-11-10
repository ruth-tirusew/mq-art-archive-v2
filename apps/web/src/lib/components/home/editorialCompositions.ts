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
	body: 'A living archive of modern Ethiopian art — portraits, practices, and the conversations that hold them.',
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

/** Enough works to fill one spread: a feature plus three supporting. */
const MIN_POOL = 4;

/**
 * One work per artist, so a spread shows a range of people. When there are too few artists to
 * fill a spread, their other works are added after, so the spreads still have something to
 * rearrange instead of repeating the same single work on every load.
 */
function uniqueByArtist(posts: ArtPost[]): ArtPost[] {
	const seen = new Set<string>();
	const out: ArtPost[] = [];
	const rest: ArtPost[] = [];
	for (const post of posts) {
		if (!hasImage(post)) continue;
		const key = artistKey(post);
		if (seen.has(key)) {
			rest.push(post);
			continue;
		}
		seen.add(key);
		out.push(post);
	}
	for (const post of rest) {
		if (out.length >= MIN_POOL) break;
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

/** A work made at least this long ago may be labelled "From the archive". */
const ARCHIVE_AGE_YEARS = 10;
/** A work published on Artiv this recently may be labelled "New on Artiv". */
const NEW_WITHIN_DAYS = 60;

export function isArchiveWork(post: ArtPost, now: Date): boolean {
	return typeof post.year === 'number' && now.getFullYear() - post.year >= ARCHIVE_AGE_YEARS;
}

export function isNewWork(post: ArtPost, now: Date): boolean {
	const published = Date.parse(post.published_at ?? '');
	if (!Number.isFinite(published) || published > now.getTime()) return false;
	return now.getTime() - published <= NEW_WITHIN_DAYS * 24 * 60 * 60 * 1000;
}

/** Takes the first post that matches out of `list` (or simply the first), so slots don't repeat. */
function takeFirst(list: ArtPost[], match: (p: ArtPost) => boolean): ArtPost | undefined {
	const i = Math.max(0, list.findIndex(match));
	return list.splice(i, 1)[0];
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
		now?: Date;
	}
): EditorialSpread[] {
	const now = opts?.now ?? new Date();
	const featuredIds = new Set((opts?.featuredPosts ?? []).map((p) => p.id));
	// A badge is a claim about the work, so it is only shown when the data backs it.
	const isFeatured = (p: ArtPost) => Boolean(p.featured_acquisition) || featuredIds.has(p.id);
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

		// The lead is a work that is actually featured when there is one. Otherwise it avoids the
		// old and the newly published works, so those can go where their label is true.
		const featurePost =
			(featuredWorks.length > 0 ? featuredWorks[i % featuredWorks.length] : undefined) ??
			(featuredSlug ? rotated.find((p) => p.artist_slug === featuredSlug) : undefined) ??
			rotated.find(isFeatured) ??
			rotated.find((p) => !isArchiveWork(p, now) && !isNewWork(p, now)) ??
			rotated[0]!;

		const rest = rotated.filter((p) => p.id !== featurePost.id);
		// Each slot prefers a work its label is true of: an old one for the archive slot, a newly
		// published one for the "new" slot. Without one, the slot still fills but goes unlabelled.
		const archivePost = takeFirst(rest, (p) => isArchiveWork(p, now)) ?? featurePost;
		const newPost = takeFirst(rest, (p) => isNewWork(p, now)) ?? featurePost;
		const supportPost = takeFirst(rest, () => true) ?? featurePost;

		spreads.push({
			layoutId,
			intro: { ...INTRO },
			feature: { post: featurePost, badge: isFeatured(featurePost) ? 'Featured' : undefined },
			support: { post: supportPost, badge: undefined },
			archive: { post: archivePost, badge: isArchiveWork(archivePost, now) ? 'From the archive' : undefined },
			emerging: { post: newPost, badge: isNewWork(newPost, now) ? 'New on Artiv' : undefined },
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
