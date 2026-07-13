import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { Artist, Work } from '$lib/data/archive';
import { postImageUrl, splitDisplayName } from '$lib/utils/display';

export { postImageUrl, splitDisplayName };

export function isArtistProfile(
	artist: ArtistProfile | Artist
): artist is ArtistProfile {
	return 'display_name' in artist;
}

export function artistName(artist: ArtistProfile | Artist): string {
	return isArtistProfile(artist) ? artist.display_name : artist.name;
}

export function artistPortrait(artist: ArtistProfile | Artist): string | undefined {
	return isArtistProfile(artist) ? artist.portrait_url : artist.portrait;
}

export function artistLocation(artist: ArtistProfile | Artist): string {
	if (isArtistProfile(artist)) return artist.contact?.location ?? '';
	return artist.based;
}

export function artistTagline(artist: ArtistProfile | Artist): string {
	if (isArtistProfile(artist)) return artist.tagline ?? artist.bio ?? '';
	return artist.tagline ?? artist.bio;
}

export function artistBorn(artist: ArtistProfile | Artist): string | undefined {
	if (isArtistProfile(artist)) return artist.born;
	return artist.born;
}

export function artistDiscipline(artist: ArtistProfile | Artist): string | undefined {
	if (isArtistProfile(artist)) return artist.discipline;
	return artist.discipline;
}

export function artistYearsActive(artist: ArtistProfile | Artist): string | undefined {
	if (isArtistProfile(artist)) return artist.years_active;
	return artist.yearsActive;
}

export function artistSlug(artist: ArtistProfile | Artist): string {
	return artist.slug;
}

export function artistHandle(artist: ArtistProfile | Artist): string {
	if (isArtistProfile(artist)) return artist.handle ?? artist.slug;
	return artist.handle ?? artist.slug;
}

export function artistInfluences(artist: ArtistProfile | Artist): string[] {
	if (isArtistProfile(artist)) return [];
	return artist.influences ?? [];
}

export function isArtPost(item: Work | ArtPost | undefined | null): item is ArtPost {
	return item != null && typeof item === 'object' && 'artist_id' in item;
}

export function acquisitionTitle(item: Work | ArtPost): string {
	return item.title;
}

export function acquisitionImage(item: Work | ArtPost): string | undefined {
	if (isArtPost(item)) return postImageUrl(item.media, item.id, item.artist_slug);
	return item.image;
}

export function acquisitionArtistSlug(item: Work | ArtPost): string {
	if (isArtPost(item)) return item.artist_slug ?? '';
	return item.artistSlug;
}

export function acquisitionArtistName(item: Work | ArtPost, fallback = ''): string {
	if (isArtPost(item)) return item.artist_name ?? fallback;
	return fallback;
}

export function acquisitionYear(item: Work | ArtPost): number | undefined {
	const year = item.year;
	return year == null ? undefined : year;
}

export function acquisitionMedium(item: Work | ArtPost): string | undefined {
	return item.medium;
}

export function acquisitionDimensions(item: Work | ArtPost): string | undefined {
	return item.dimensions;
}

export function acquisitionPalette(item: Work | ArtPost): string[] {
	return item.palette ?? [];
}

export function acquisitionFeatured(item: Work | ArtPost): boolean {
	if (isArtPost(item)) return Boolean(item.featured_acquisition);
	return Boolean(item.featuredAcquisition);
}
