import type { PublicAnswer } from '$lib/core/domain/questionnaire';

export const fixturePublicAnswers: PublicAnswer[] = [
	{
		key: 'bio',
		type: 'long_text',
		prompt_en: 'Tell us about your practice',
		prompt_am: 'ስለ ስራዎ ይንገሩን',
		body: '<p>ሰላም <strong>ዓለም</strong></p>',
		format: 'rich',
		lang: 'am'
	},
	{ key: 'city', type: 'short_text', prompt_en: 'City', body: 'Addis Ababa', format: 'plain', lang: 'en' },
	{ key: 'site', type: 'link', prompt_en: 'Website', body: 'https://example.com/work', format: 'plain', lang: 'en' }
];
