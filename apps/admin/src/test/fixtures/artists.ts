import type { ArtistProfile } from '$lib/core/domain/artist';

export const fixtureArtist: ArtistProfile = {
	id: 'artist-1',
	slug: 'selamawit-tesfaye',
	handle: 'selam',
	display_name: 'Selamawit Tesfaye',
	discipline: 'Painting',
	featured: true,
	status: 'approved'
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
		status: 'pending'
	}
];
