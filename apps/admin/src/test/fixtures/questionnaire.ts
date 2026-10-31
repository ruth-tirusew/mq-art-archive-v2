import type { Invite, InviteCreated } from '$lib/core/domain/invite';
import type { NudgePreview, QuestionnaireRevision, ReviewView } from '$lib/core/domain/questionnaire';

export const fixtureRevision: QuestionnaireRevision = {
	id: 'rev-1',
	number: 1,
	status: 'published',
	change_kind: 'substantive',
	published_at: '2026-10-01T00:00:00Z',
	created_at: '2026-10-01T00:00:00Z',
	updated_at: '2026-10-01T00:00:00Z',
	questions: [
		{ key: 'bio', position: 1, type: 'long_text', prompt_en: 'Tell us about your practice', prompt_am: 'ስለ ስራዎ ይንገሩን', required: true, is_public: true, max_length: 5000 },
		{ key: 'income', position: 2, type: 'short_text', prompt_en: 'Income range', required: false, is_public: false, max_length: 200 }
	]
};

export const fixtureDraft: QuestionnaireRevision = {
	...fixtureRevision,
	id: 'rev-draft',
	number: undefined,
	status: 'draft',
	published_at: undefined
};

export const fixtureNudgePreview: NudgePreview = {
	revision_id: 'rev-1',
	revision_number: 1,
	change_kind: 'substantive',
	eligible: 3,
	can_send: true
};

export const fixtureReviewView: ReviewView = {
	revision: fixtureRevision,
	items: [
		{
			question: fixtureRevision.questions[0],
			answer: { key: 'bio', body: '<p>ሰላም</p>', format: 'rich', lang: 'am', updated_at: '2026-10-02T00:00:00Z' }
		},
		{ question: fixtureRevision.questions[1], answer: null }
	],
	other: []
};

export const fixtureInvite: Invite = {
	id: 'inv-1',
	email: 'ada@example.com',
	display_name: 'Ada',
	requested_handle: 'ada_1',
	status: 'live',
	expires_at: '2026-11-14T00:00:00Z',
	created_at: '2026-10-31T00:00:00Z'
};

export const fixtureInviteCreated: InviteCreated = {
	invite: fixtureInvite,
	token: 'tok_abc',
	link: 'https://artiv.test/apply?invite=tok_abc'
};
