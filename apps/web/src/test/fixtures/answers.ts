import type { PublicAnswer, QuestionnaireForm, QuestionnaireNudge } from '$lib/core/domain/questionnaire';

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

export const fixtureForm: QuestionnaireForm = {
	revision: {
		id: 'rev-1',
		number: 1,
		questions: [
			{ key: 'bio', position: 1, type: 'long_text', prompt_en: 'Tell us about your practice', required: true, is_public: true, max_length: 5000 },
			{ key: 'income', position: 2, type: 'short_text', prompt_en: 'Income range', required: false, is_public: false, max_length: 200 }
		]
	},
	answers: [{ key: 'bio', body: '<p>Hello</p>', format: 'rich', lang: 'en', updated_at: '2026-10-01T00:00:00Z' }],
	locked: false
};

export const fixtureNudge: QuestionnaireNudge = {
	revision_id: 'rev-2',
	sent_at: '2026-10-31T00:00:00Z',
	questions: [{ key: 'influences', position: 3, type: 'long_text', prompt_en: 'Influences', required: false, is_public: true, max_length: 5000 }]
};
