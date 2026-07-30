import { describe, expect, expectTypeOf, it } from 'vitest';
import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';
import {
	acquisitionArtistSlug,
	acquisitionImage,
	artistName,
	artistPortrait
} from './fields';

describe('API field helpers', () => {
	it('accepts only API domain shapes', () => {
		expectTypeOf(artistName).parameter(0).toEqualTypeOf<ArtistProfile>();
		expectTypeOf(artistPortrait).parameter(0).toEqualTypeOf<ArtistProfile>();
		expectTypeOf(acquisitionImage).parameter(0).toEqualTypeOf<ArtPost>();
		expectTypeOf(acquisitionArtistSlug).parameter(0).toEqualTypeOf<ArtPost>();
	});

	it('reads normalized API fields', () => {
		const artist = {
			id: 'artist-1',
			slug: 'selam',
			display_name: 'Selam',
			portrait_url: 'https://example.test/portrait.jpg'
		} satisfies ArtistProfile;
		const post = {
			id: 'post-1',
			artist_id: artist.id,
			artist_slug: artist.slug,
			title: 'Blue',
			media: [{ id: 'media-1', url: 'https://example.test/blue.jpg' }]
		} satisfies ArtPost;

		expect(artistName(artist)).toBe('Selam');
		expect(artistPortrait(artist)).toBe(artist.portrait_url);
		expect(acquisitionArtistSlug(post)).toBe('selam');
		expect(acquisitionImage(post)).toBe(post.media[0].url);
	});
});
