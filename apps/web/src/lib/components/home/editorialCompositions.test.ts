import { describe, expect, it } from 'vitest';
import type { ArtPost } from '$lib/core/domain/art';
import { buildEditorialSpreads } from './editorialCompositions';

const post = (id: string, artist: string): ArtPost =>
	({
		id,
		artist_slug: artist,
		artist_id: artist,
		title: id,
		media: [{ id, url: `https://img.test/${id}.jpg`, sort_order: 0 }]
	}) as unknown as ArtPost;

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
