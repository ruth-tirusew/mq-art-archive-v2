export type WikiArticle = {
	slug: string;
	title: string;
	category: 'Legal' | 'Materials' | 'Pricing' | 'Contracts' | 'Distribution';
	excerpt: string;
	verified: boolean;
	reviewers: number;
	updated: string;
	contributors: number;
	readingTime: number;
	difficulty: 'Beginner' | 'Intermediate' | 'Advanced';
};

export const wikiArticles: WikiArticle[] = [
	{
		slug: 'eipa-registration',
		title: 'Registering your work with the EIPA',
		category: 'Legal',
		excerpt:
			'Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.',
		verified: true,
		reviewers: 4,
		updated: '2026-05-14',
		contributors: 11,
		readingTime: 8,
		difficulty: 'Beginner'
	},
	{
		slug: 'addis-pigment-sources',
		title: 'Where to buy pigments, linen and stretchers in Addis',
		category: 'Materials',
		excerpt:
			'A working list of Piassa, Mercato and Bole suppliers — what they stock, who imports, and price ranges.',
		verified: true,
		reviewers: 6,
		updated: '2026-06-02',
		contributors: 23,
		readingTime: 12,
		difficulty: 'Beginner'
	},
	{
		slug: 'pricing-local-vs-international',
		title: 'Pricing commissions: local vs. international clients',
		category: 'Pricing',
		excerpt:
			'How to set a base rate in ETB and USD, account for FX volatility, and avoid common underpricing traps.',
		verified: false,
		reviewers: 1,
		updated: '2026-06-18',
		contributors: 7,
		readingTime: 15,
		difficulty: 'Intermediate'
	},
	{
		slug: 'gallery-contracts',
		title: 'Reading a gallery consignment contract',
		category: 'Contracts',
		excerpt:
			'Commission splits, insurance, exclusivity windows, and the clauses Ethiopian artists keep getting burned by.',
		verified: true,
		reviewers: 3,
		updated: '2026-04-30',
		contributors: 9,
		readingTime: 18,
		difficulty: 'Advanced'
	},
	{
		slug: 'shipping-works-abroad',
		title: 'Shipping works abroad from Addis',
		category: 'Distribution',
		excerpt:
			'Crating, customs paperwork, DHL vs. freight forwarders, and what collectors actually pay for.',
		verified: false,
		reviewers: 2,
		updated: '2026-06-20',
		contributors: 5,
		readingTime: 10,
		difficulty: 'Beginner'
	},
	{
		slug: 'fair-use-amharic',
		title: 'Fair use and licensing in an Ethiopian context',
		category: 'Legal',
		excerpt:
			'What Ethiopian copyright law actually says about derivative works, sampling, and reference photography.',
		verified: true,
		reviewers: 5,
		updated: '2026-03-21',
		contributors: 14,
		readingTime: 14,
		difficulty: 'Advanced'
	}
];

export const wikiCategories = ['All', 'Legal', 'Materials', 'Pricing', 'Contracts', 'Distribution'] as const;

export const listWikiArticles = () => wikiArticles;

export const getWikiArticle = (slug: string) => wikiArticles.find((a) => a.slug === slug);
