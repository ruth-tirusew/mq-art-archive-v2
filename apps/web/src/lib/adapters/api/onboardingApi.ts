import { apiFetch } from '$lib/adapters/api/client';
import type {
	HandleAvailability,
	OnboardingApplication,
	SubmitApplicationInput
} from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export class OnboardingApi implements OnboardingPort {
	getMyApplication(): Promise<OnboardingApplication> {
		return apiFetch<OnboardingApplication>('/api/v1/applications/me');
	}

	checkHandleAvailability(handle: string): Promise<HandleAvailability> {
		return apiFetch<HandleAvailability>(
			`/api/v1/handles/${encodeURIComponent(handle)}/available`
		);
	}

	submit(input: SubmitApplicationInput): Promise<OnboardingApplication> {
		return apiFetch<OnboardingApplication>('/api/v1/applications', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		});
	}
}
