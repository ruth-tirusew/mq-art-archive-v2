import type { ArtistProfile } from '$lib/core/domain/profile';

export const fixtureArtist: ArtistProfile = {
	id: 'artist-1',
	slug: 'selamawit-tesfaye',
	handle: 'selam',
	display_name: 'Selamawit Tesfaye',
	bio: 'Painter based in Addis Ababa.',
	discipline: 'Painting',
	tagline: 'Color as memory',
	portrait_url: 'https://example.com/portrait.jpg',
	featured: true,
	status: 'approved',
	contact: { location: 'Addis Ababa' }
};

export const fixtureArtists: ArtistProfile[] = [
	fixtureArtist,
	{
		id: 'artist-2',
		slug: 'yonas-bekele',
		handle: 'yonas',
		display_name: 'Yonas Bekele',
		discipline: 'Photography',
		featured: false,
		status: 'approved'
	}
];
