import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';
import { postImageUrl, splitDisplayName } from '$lib/utils/display';

export { postImageUrl, splitDisplayName };

export function artistName(artist: ArtistProfile): string {
	return artist.display_name;
}

export function artistPortrait(artist: ArtistProfile): string | undefined {
	return artist.portrait_url;
}

export function artistLocation(artist: ArtistProfile): string {
	return artist.contact?.location ?? '';
}

export function artistTagline(artist: ArtistProfile): string {
	return artist.tagline ?? artist.bio ?? '';
}

export function artistBorn(artist: ArtistProfile): string | undefined {
	return artist.born;
}

export function artistDiscipline(artist: ArtistProfile): string | undefined {
	return artist.discipline;
}

export function artistYearsActive(artist: ArtistProfile): string | undefined {
	return artist.years_active;
}

export function artistSlug(artist: ArtistProfile): string {
	return artist.slug;
}

export function artistHandle(artist: ArtistProfile): string {
	return artist.handle ?? artist.slug;
}

export function artistInfluences(artist: ArtistProfile): string[] {
	return artist.influences ?? [];
}

export function artistInResidence(artist: ArtistProfile): boolean {
	return artist.in_residence ?? false;
}

export function artistResidencyPlace(artist: ArtistProfile): string | undefined {
	return artist.residency_place || undefined;
}

export function artistOpenForCommission(artist: ArtistProfile): boolean {
	return artist.open_for_commission ?? false;
}

export function acquisitionTitle(item: ArtPost): string {
	return item.title;
}

export function acquisitionImage(item: ArtPost): string | undefined {
	return postImageUrl(item.media, item.id, item.artist_slug);
}

export function acquisitionArtistSlug(item: ArtPost): string {
	return item.artist_slug ?? '';
}

export function acquisitionArtistName(item: ArtPost, fallback = ''): string {
	return item.artist_name ?? fallback;
}

export function acquisitionYear(item: ArtPost): number | undefined {
	const year = item.year;
	return year == null ? undefined : year;
}

export function acquisitionMedium(item: ArtPost): string | undefined {
	return item.medium;
}

export function acquisitionDimensions(item: ArtPost): string | undefined {
	return item.dimensions;
}

export function acquisitionPalette(item: ArtPost): string[] {
	return item.palette ?? [];
}

export function acquisitionFeatured(item: ArtPost): boolean {
	return Boolean(item.featured_acquisition);
}
