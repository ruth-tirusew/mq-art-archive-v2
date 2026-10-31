import { apiFetch } from '$lib/adapters/api/client';
import type { ClaimInviteInput, InvitePreview } from '$lib/core/domain/invite';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';

const jsonPost = (body: unknown): RequestInit => ({
	method: 'POST',
	headers: { 'Content-Type': 'application/json' },
	body: JSON.stringify(body)
});

export class InviteApi {
	/** The token goes in the body, never the URL, so it stays out of logs and history. */
	preview(token: string): Promise<InvitePreview> {
		return apiFetch<InvitePreview>('/api/v1/invites/preview', jsonPost({ token }));
	}

	/** Claims the invite and submits the application in one step. */
	claim(input: ClaimInviteInput): Promise<OnboardingApplication> {
		return apiFetch<OnboardingApplication>('/api/v1/applications/invite', jsonPost(input));
	}
}
