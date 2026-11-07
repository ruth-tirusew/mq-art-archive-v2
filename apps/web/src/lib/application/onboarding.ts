import type {
	HandleAvailability,
	OnboardingApplication,
	SubmitApplicationInput
} from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export function createOnboardingService(port: OnboardingPort) {
	return {
		getMyApplication(): Promise<OnboardingApplication> {
			return port.getMyApplication();
		},
		checkHandleAvailability(handle: string): Promise<HandleAvailability> {
			return port.checkHandleAvailability(handle);
		},
		submit(input: SubmitApplicationInput): Promise<OnboardingApplication> {
			return port.submit(input);
		}
	};
}

export type OnboardingService = ReturnType<typeof createOnboardingService>;
