import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';

/**
 * Presentation only, not product content (which always comes from the API): a made-up profile
 * shown wherever a real artist cannot be: the /portfolio preview before anyone
 * is featured, or while the API is down. It is a template ("Your Name"), never a real person, so
 * its contact buttons are inert (see ShareableProfile's `sample`) and its images are drawn here
 * instead of loaded from anywhere.
 */

/** A soft, watercolour-like swatch as an SVG data URI: no file, no network, no copyright. */
function swatch(from: string, to: string, accent: string): string {
	const svg =
		`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 400" preserveAspectRatio="xMidYMid slice">` +
		`<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${from}"/><stop offset="1" stop-color="${to}"/></linearGradient>` +
		`<filter id="b"><feGaussianBlur stdDeviation="28"/></filter></defs>` +
		`<rect width="400" height="400" fill="url(#g)"/>` +
		`<circle cx="130" cy="150" r="110" fill="${accent}" opacity="0.55" filter="url(#b)"/>` +
		`<circle cx="290" cy="270" r="90" fill="${to}" opacity="0.7" filter="url(#b)"/></svg>`;
	return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

export const SAMPLE_ARTIST: ArtistProfile = {
	id: 'sample-artist',
	slug: 'yourname',
	handle: 'yourname',
	display_name: 'Your Name',
	discipline: 'Painter',
	tagline: 'A line about the work you make.',
	bio: '<p>A few lines about you and your work. Visitors read this first, so say what you make and what it is about.</p>',
	open_for_commission: true,
	contact: { location: 'Addis Ababa' },
	social: {},
	status: 'approved'
} as ArtistProfile;

const work = (n: number, title: string, from: string, to: string, accent: string): ArtPost => ({
	id: `sample-work-${n}`,
	artist_id: SAMPLE_ARTIST.id,
	artist_slug: SAMPLE_ARTIST.slug,
	title,
	status: 'published',
	media: [{ id: `sample-media-${n}`, url: swatch(from, to, accent), sort_order: 0 }]
});

export const SAMPLE_WORKS: ArtPost[] = [
	work(1, 'Your first piece', '#f3d9b8', '#c9774a', '#e9a66a'),
	work(2, 'Your second piece', '#cfe0d6', '#4f7c6f', '#8fb8a5'),
	work(3, 'Your third piece', '#e6d3ea', '#7a5c9c', '#b79bd0')
];
