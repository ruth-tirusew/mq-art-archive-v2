export type InviteStatus = 'live' | 'claimed' | 'revoked' | 'expired';

/** What the claim page may show before sign-in. */
export interface InvitePreview {
	email: string;
	display_name?: string;
	handle?: string;
	status: InviteStatus;
}

export interface ClaimInviteInput {
	token: string;
	display_name?: string;
	requested_handle?: string;
	notes?: string;
}
