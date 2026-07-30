import type {
	HandleAvailability,
	OnboardingApplication,
	SubmitApplicationInput
} from '$lib/core/domain/onboarding';

export interface OnboardingPort {
	getMyApplication(): Promise<OnboardingApplication>;
	checkHandleAvailability(handle: string): Promise<HandleAvailability>;
	submit(input: SubmitApplicationInput): Promise<OnboardingApplication>;
}
