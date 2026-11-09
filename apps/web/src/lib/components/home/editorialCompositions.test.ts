import { describe, expect, it } from 'vitest';
import type { ArtPost } from '$lib/core/domain/art';
import { buildEditorialSpreads, isArchiveWork, isNewWork } from './editorialCompositions';

const post = (id: string, artist: string, over: Partial<ArtPost> = {}): ArtPost =>
	({
		id,
		artist_slug: artist,
		artist_id: artist,
		title: id,
		media: [{ id, url: `https://img.test/${id}.jpg`, sort_order: 0 }],
		...over
	}) as unknown as ArtPost;

const now = new Date('2026-11-10T12:00:00Z');
const daysAgo = (d: number) => new Date(now.getTime() - d * 86_400_000).toISOString();

const featureIds = (seed: number, posts: ArtPost[]) =>
	buildEditorialSpreads(posts, { spreadCount: 4, seed }).map((s) => s.feature.post.id);

describe('buildEditorialSpreads', () => {
	it('still rearranges when one artist has all the works', () => {
		const posts = ['a', 'b', 'c', 'd', 'e', 'f'].map((id) => post(id, 'solo'));
		const seen = new Set(Array.from({ length: 12 }, (_, seed) => featureIds(seed, posts).join()));
		expect(seen.size).toBeGreaterThan(1);
	});

	it('keeps one work per artist when there are enough artists', () => {
		const posts = [post('a1', 'a'), post('a2', 'a'), post('b1', 'b'), post('c1', 'c'), post('d1', 'd')];
		const spread = buildEditorialSpreads(posts, { spreadCount: 1, seed: 1 })[0]!;
		const shown = [spread.feature.post, spread.support.post, spread.archive.post, spread.emerging.post];
		expect(new Set(shown.map((p) => p.artist_slug)).size).toBe(shown.length);
	});
});

describe('spread badges only say what is true', () => {
	it('labels nothing when no work qualifies', () => {
		const posts = ['a', 'b', 'c', 'd'].map((id, i) => post(id, `artist${i}`, { year: 2024, published_at: daysAgo(400) }));
		for (const s of buildEditorialSpreads(posts, { spreadCount: 3, seed: 7, now })) {
			expect(s.feature.badge).toBeUndefined();
			expect(s.archive.badge).toBeUndefined();
			expect(s.emerging.badge).toBeUndefined();
		}
	});

	it('puts a qualifying work in each slot and labels it', () => {
		const posts = [
			post('old', 'a', { year: 1964, published_at: daysAgo(400) }),
			post('fresh', 'b', { year: 2025, published_at: daysAgo(5) }),
			post('picked', 'c', { year: 2023, published_at: daysAgo(400), featured_acquisition: true }),
			post('plain', 'd', { year: 2023, published_at: daysAgo(400) })
		];
		for (const s of buildEditorialSpreads(posts, { spreadCount: 4, seed: 3, now })) {
			expect(s.archive.post.id).toBe('old');
			expect(s.archive.badge).toBe('From the archive');
			expect(s.emerging.post.id).toBe('fresh');
			expect(s.emerging.badge).toBe('New on Artiv');
			expect(s.feature.badge).toBe(s.feature.post.featured_acquisition ? 'Featured' : undefined);
		}
	});

	it('judges archive by the year made and "new" by the date published', () => {
		expect(isArchiveWork(post('x', 'a', { year: 2016 }), now)).toBe(true);
		expect(isArchiveWork(post('x', 'a', { year: 2017 }), now)).toBe(false);
		expect(isArchiveWork(post('x', 'a'), now)).toBe(false);
		expect(isNewWork(post('x', 'a', { published_at: daysAgo(59) }), now)).toBe(true);
		expect(isNewWork(post('x', 'a', { published_at: daysAgo(61) }), now)).toBe(false);
		expect(isNewWork(post('x', 'a', { published_at: null }), now)).toBe(false);
	});
});
