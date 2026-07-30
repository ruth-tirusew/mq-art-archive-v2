import type { ArtPost } from '$lib/core/domain/art';

export const fixturePost: ArtPost = {
	id: 'post-1',
	artist_id: 'artist-1',
	artist_slug: 'selamawit-tesfaye',
	artist_name: 'Selamawit Tesfaye',
	title: 'Blue Market',
	description: 'Oil on canvas',
	medium: 'Oil',
	year: 2024,
	city: 'Addis Ababa',
	status: 'published',
	media: [{ id: 'm1', url: 'https://example.com/work.jpg' }]
};

export const fixturePosts: ArtPost[] = [fixturePost];
