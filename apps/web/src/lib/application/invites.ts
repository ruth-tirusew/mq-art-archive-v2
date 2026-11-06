import { InviteApi } from '$lib/adapters/api/inviteApi';
import type { User } from '$lib/core/domain/auth';
import type { ClaimInviteInput, InvitePreview } from '$lib/core/domain/invite';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';

const api = new InviteApi();

export const inviteService = {
	preview(token: string): Promise<InvitePreview> {
		return api.preview(token);
	},
	register(token: string, password: string): Promise<User> {
		return api.register(token, password);
	},
	claim(input: ClaimInviteInput): Promise<OnboardingApplication> {
		return api.claim(input);
	}
};
