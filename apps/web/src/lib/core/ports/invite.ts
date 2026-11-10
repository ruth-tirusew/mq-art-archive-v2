import type { User } from '$lib/core/domain/auth';
import type { ClaimInviteInput, InvitePreview } from '$lib/core/domain/invite';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';

export interface InvitePort {
	preview(token: string): Promise<InvitePreview>;
	register(token: string, password: string): Promise<User>;
	claim(input: ClaimInviteInput): Promise<OnboardingApplication>;
}
