/** Legacy static shapes kept for type unions with API domain models. */

export type Work = {
	id: string;
	title: string;
	year: number;
	medium: string;
	dimensions?: string;
	image: string;
	artistSlug: string;
	palette: string[];
	city: string;
	style: string;
	residency?: string;
	exhibition?: string;
	featuredAcquisition?: boolean;
};

export type TimelineEntry = {
	year: number;
	kind: 'birth' | 'study' | 'exhibition' | 'residency' | 'work' | 'note';
	title: string;
	place?: string;
	detail?: string;
};

export type IdentityMarker = {
	label: string;
	tone: 'accent' | 'muted' | 'secondary' | 'highlight';
};

export type ArtistLinks = {
	telegram?: string;
	instagram?: string;
	whatsapp?: string;
};

export type Artist = {
	slug: string;
	handle?: string;
	name: string;
	born: string;
	based: string;
	discipline: string;
	portrait: string;
	bio: string;
	influences: string[];
	featured?: boolean;
	verified?: boolean;
	tagline?: string;
	yearsActive?: string;
	timeline?: TimelineEntry[];
	identityMarkers?: IdentityMarker[];
	links?: ArtistLinks;
	demo?: boolean;
};
