import type { User } from '$lib/core/domain/auth';
import type { ClaimInviteInput, InvitePreview } from '$lib/core/domain/invite';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';
import type { InvitePort } from '$lib/core/ports/invite';

export function createInviteService(port: InvitePort) {
	return {
		preview(token: string): Promise<InvitePreview> {
			return port.preview(token);
		},
		register(token: string, password: string): Promise<User> {
			return port.register(token, password);
		},
		claim(input: ClaimInviteInput): Promise<OnboardingApplication> {
			return port.claim(input);
		}
	};
}

export type InviteService = ReturnType<typeof createInviteService>;
