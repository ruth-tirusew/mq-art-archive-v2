import { apiFetch } from '$lib/adapters/api/client';
import type { User } from '$lib/core/domain/auth';
import type { ClaimInviteInput, InvitePreview } from '$lib/core/domain/invite';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';
import type { InvitePort } from '$lib/core/ports/invite';

const jsonPost = (body: unknown): RequestInit => ({
	method: 'POST',
	headers: { 'Content-Type': 'application/json' },
	body: JSON.stringify(body)
});

export class InviteApi implements InvitePort {
	/** The token goes in the body, never the URL, so it stays out of logs and history. */
	preview(token: string): Promise<InvitePreview> {
		return apiFetch<InvitePreview>('/api/v1/invites/preview', jsonPost({ token }));
	}

	/**
	 * Creates the invitee's account straight from the invite and signs them in (sets the session
	 * cookie). Only the token and a password are sent: the account's email is the invite's, decided
	 * by the server. It is created already verified, since the link was sent only to that inbox.
	 */
	register(token: string, password: string): Promise<User> {
		return apiFetch<User>('/api/v1/invites/register', jsonPost({ token, password }));
	}

	/** Claims the invite and submits the application in one step. */
	claim(input: ClaimInviteInput): Promise<OnboardingApplication> {
		return apiFetch<OnboardingApplication>('/api/v1/applications/invite', jsonPost(input));
	}
}
