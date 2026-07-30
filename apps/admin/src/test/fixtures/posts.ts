import type { ArtPost } from '$lib/core/domain/art';

export const fixturePost: ArtPost = {
	id: 'post-1',
	artist_id: 'artist-1',
	artist_slug: 'selamawit-tesfaye',
	artist_name: 'Selamawit Tesfaye',
	title: 'Blue Market',
	status: 'published',
	featured_acquisition: false,
	media: [{ url: 'https://example.com/work.jpg' }]
};

export const fixturePosts: ArtPost[] = [fixturePost];
