import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import type {
	HandleAvailability,
	OnboardingApplication,
	SubmitApplicationInput
} from '$lib/core/domain/onboarding';

const api = new OnboardingApi();

export const onboardingService = {
	getMyApplication(): Promise<OnboardingApplication> {
		return api.getMyApplication();
	},
	checkHandleAvailability(handle: string): Promise<HandleAvailability> {
		return api.checkHandleAvailability(handle);
	},
	submit(input: SubmitApplicationInput): Promise<OnboardingApplication> {
		return api.submit(input);
	}
};
