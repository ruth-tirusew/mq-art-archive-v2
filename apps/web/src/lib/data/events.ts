export type Event = {
	id: string;
	slug: string;
	title: string;
	description: string;
	venue: string;
	city: string;
	date: string;
	time?: string;
	type: 'Opening' | 'Talk' | 'Poetry' | 'Theatre' | 'Pop-up' | 'Screening';
	source: 'Scraped' | 'Institutional' | 'Submitted';
	host?: string;
	externalUrl?: string;
	relatedArtistSlugs?: string[];
};

export const TELEGRAM_DIGEST_URL = 'https://t.me/makdas_events';

export const events: Event[] = [
	{
		id: 'e1',
		slug: 'after-the-rains-tewodros-hailu',
		title: 'After the Rains — Tewodros Hailu solo',
		description:
			"Tewodros Hailu's first solo at Addis Fine Art in three years — nine new pigment-soaked linens hung unstretched, pinned like drying laundry. The work circles residue: walls, weather, and the memory of both.\n\nOpening night includes a walkthrough with the artist at 19:00. Works remain on view through August.",
		venue: 'Addis Fine Art',
		city: 'Addis Ababa',
		date: '2026-06-27',
		time: '18:00 EAT',
		type: 'Opening',
		source: 'Institutional',
		host: 'Addis Fine Art',
		externalUrl: 'https://addisfineart.com',
		relatedArtistSlugs: ['tewodros-hailu']
	},
	{
		id: 'e2',
		slug: 'tobiya-poetic-jazz-night',
		title: 'Tobiya Poetic Jazz Night',
		description:
			'A monthly evening of Amharic poetry, live jazz, and open mic at Fendika. This month\'s theme: migration and return.\n\nDoors at 20:00. Entry 150 ETB at the door.',
		venue: 'Fendika Cultural Center',
		city: 'Addis Ababa',
		date: '2026-06-28',
		time: '20:00 EAT',
		type: 'Poetry',
		source: 'Scraped'
	},
	{
		id: 'e3',
		slug: 'curators-walkthrough-skunders-cosmos',
		title: "Curator's Walkthrough — Skunder's Cosmos",
		description:
			'A guided tour of the Skunder Boghossian retrospective with Alle School faculty. Focus on diasporic dream-spaces and the artist\'s chromatic grammar.\n\nFree with museum admission. Meet in the main hall.',
		venue: 'Modern Art Museum / Gebre Kristos Desta Center',
		city: 'Addis Ababa',
		date: '2026-07-02',
		time: '15:00 EAT',
		type: 'Talk',
		source: 'Institutional',
		host: 'Alle School of Fine Arts',
		externalUrl: 'https://mamgkdc.org'
	},
	{
		id: 'e4',
		slug: 'design-pop-up-habesha-futures',
		title: 'Design Pop-up: Habesha Futures',
		description:
			'A one-weekend pop-up at Zoma Museum featuring textile designers, furniture makers, and illustrators reimagining Ethiopian craft for contemporary interiors.\n\nSaturday and Sunday, 10:00–18:00.',
		venue: 'Zoma Museum',
		city: 'Addis Ababa',
		date: '2026-07-05',
		time: '10:00 EAT',
		type: 'Pop-up',
		source: 'Submitted',
		externalUrl: 'https://zomamuseum.org'
	},
	{
		id: 'e5',
		slug: 'yenegew-sew-theatre-revival',
		title: 'Yenegew Sew — Theatre revival',
		description:
			'A revival of the classic Ethiopian play at the National Theatre — directed by a new generation of Alle School graduates. Subtitles in English for select performances.\n\nEvening shows Thursday through Sunday.',
		venue: 'National Theatre',
		city: 'Addis Ababa',
		date: '2026-07-09',
		time: '19:30 EAT',
		type: 'Theatre',
		source: 'Scraped'
	},
	{
		id: 'e6',
		slug: 'goethe-film-series-diaspora-editions',
		title: 'Goethe Film Series: Diaspora Editions',
		description:
			'Monthly screening of contemporary African and diaspora cinema, followed by a moderated discussion. This edition features work from Ethiopian filmmakers in Berlin and London.\n\nSeating is limited — arrive early.',
		venue: 'Goethe-Institut',
		city: 'Addis Ababa',
		date: '2026-07-11',
		time: '18:30 EAT',
		type: 'Screening',
		source: 'Institutional',
		host: 'Goethe-Institut',
		externalUrl: 'https://goethe.de/addis'
	},
	{
		id: 'e7',
		slug: 'bahir-dar-lake-sessions-open-studios',
		title: 'Bahir Dar Lake Sessions — Open studios',
		description:
			'Open studios along Lake Tana with painters, weavers, and ceramicists working in a shared boathouse space. A chance to meet artists outside the Addis circuit.\n\nFree entry. Works available for direct purchase from studios.',
		venue: 'Lake Tana Boathouse',
		city: 'Bahir Dar',
		date: '2026-07-13',
		time: '11:00 EAT',
		type: 'Pop-up',
		source: 'Submitted'
	}
];

export const eventFilters = ['All', 'Opening', 'Talk', 'Poetry', 'Theatre', 'Pop-up', 'Screening'] as const;

export const getEvent = (slug: string) => events.find((e) => e.slug === slug);
export const getEventById = (id: string) => events.find((e) => e.id === id);
