export type InviteStatus = 'live' | 'claimed' | 'revoked' | 'expired';

export interface Invite {
  id: string;
  email: string;
  display_name?: string;
  requested_handle?: string;
  status: InviteStatus;
  expires_at: string;
  claimed_at?: string;
  revoked_at?: string;
  created_at: string;
}

/** Returned once, on create or resend. The token and link cannot be retrieved again. */
export interface InviteCreated {
  invite: Invite;
  token: string;
  link: string;
}

export interface CreateInviteInput {
  email: string;
  display_name?: string;
  handle?: string;
  /** 0 or omitted means the server default (14 days). */
  ttl_days?: number;
}
