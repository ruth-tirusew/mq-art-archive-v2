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

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const HANDLE = /^[a-z0-9_]{3,30}$/;

/** Handles are case-insensitive; the API stores them lowercase, so that is what gets sent. */
export function normalizeInviteHandle(raw: string): string {
  return raw.trim().toLowerCase();
}

export interface InviteFormErrors {
  email?: string;
  handle?: string;
}

/** Mirrors the API's checks so the form can answer before a round trip; the server stays authoritative. */
export function validateInviteForm(input: { email: string; handle: string }): InviteFormErrors {
  const errors: InviteFormErrors = {};
  if (!EMAIL.test(input.email.trim())) errors.email = 'Enter a valid email address.';
  const handle = normalizeInviteHandle(input.handle);
  if (handle && !HANDLE.test(handle)) errors.handle = 'Use 3–30 letters, numbers or underscores.';
  return errors;
}

export const inviteStatusLabel: Record<InviteStatus, string> = {
  live: 'Waiting',
  claimed: 'Claimed',
  expired: 'Expired',
  revoked: 'Revoked'
};

/** A used invite can't be sent again; anything else can, which replaces the old link. */
export function canResendInvite(invite: Pick<Invite, 'status'>): boolean {
  return invite.status !== 'claimed';
}

/** Only a waiting invite has anything left to cancel. */
export function canRevokeInvite(invite: Pick<Invite, 'status'>): boolean {
  return invite.status === 'live';
}
