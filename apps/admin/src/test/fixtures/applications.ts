import type { OnboardingApplication } from '$lib/core/domain/onboarding';

export const fixtureApplication: OnboardingApplication = {
	id: 'app-1',
	applicant_type: 'artist',
	status: 'pending',
	display_name: 'New Applicant',
	notes: 'Looking to join the archive.',
	created_at: '2024-01-01T00:00:00Z',
	updated_at: '2024-01-01T00:00:00Z'
};

export const fixtureApplications: OnboardingApplication[] = [fixtureApplication];
