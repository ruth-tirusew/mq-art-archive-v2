export type StaffInviteStatus = 'live' | 'claimed' | 'revoked' | 'expired';

export interface StaffInvite {
  id: string;
  email: string;
  role: 'moderator';
  status: StaffInviteStatus;
  expires_at: string;
  claimed_at?: string;
  created_at: string;
}

/** Returned once, on create or resend. The link cannot be retrieved again: only a hash is stored. */
export interface StaffInviteCreated {
  invite: StaffInvite;
  link: string;
}

export interface StaffInvitePage {
  items: StaffInvite[];
  /** Matches across all pages, not just this one. */
  total: number;
}

/** What the accept page shows before the invitee picks a password. */
export interface StaffInvitePreview {
  email: string;
  role: 'moderator';
  status: StaffInviteStatus;
  expires_at: string;
}

export const STAFF_INVITES_PAGE_SIZE = 50;

export const staffInviteStatusLabel: Record<StaffInviteStatus, string> = {
  live: 'Waiting',
  claimed: 'Accepted',
  expired: 'Expired',
  revoked: 'Revoked'
};

/** A used invite can't be sent again; anything else can, which replaces the old link. */
export function canResendStaffInvite(invite: Pick<StaffInvite, 'status'>): boolean {
  return invite.status !== 'claimed';
}

/** Only a waiting invite has anything left to cancel. */
export function canRevokeStaffInvite(invite: Pick<StaffInvite, 'status'>): boolean {
  return invite.status === 'live';
}

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

/** The reason an address can't be invited, or null. The server stays authoritative. */
export function staffInviteEmailProblem(email: string): string | null {
  return EMAIL.test(email.trim()) ? null : 'Enter a valid email address.';
}

export const MIN_PASSWORD_LENGTH = 8;
/** The API hashes with bcrypt, which only reads the first 72 BYTES: an Ethiopic letter is 3 bytes. */
export const MAX_PASSWORD_BYTES = 72;

/** The reason a password can't be used, or null. Counts bytes, like the server. */
export function passwordProblem(password: string): string | null {
  if (password.length < MIN_PASSWORD_LENGTH) return `Use at least ${MIN_PASSWORD_LENGTH} characters.`;
  if (new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES) {
    return `That password is too long (at most ${MAX_PASSWORD_BYTES} bytes; Amharic letters count as 3 each).`;
  }
  return null;
}
