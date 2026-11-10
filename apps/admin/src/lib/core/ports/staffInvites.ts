import type { StaffInviteCreated, StaffInvitePage, StaffInvitePreview } from '$lib/core/domain/staffInvite';
import type { User } from '$lib/core/domain/auth';

export interface StaffInvitesPort {
  list(page: { limit: number; offset: number }): Promise<StaffInvitePage>;
  create(email: string): Promise<StaffInviteCreated>;
  resend(id: string): Promise<StaffInviteCreated>;
  revoke(id: string): Promise<void>;
  /** Public: used by the accept page before anyone is signed in. */
  preview(token: string): Promise<StaffInvitePreview>;
  /** Public: creates the account and signs the invitee in (the session cookie is set by the API). */
  accept(token: string, password: string): Promise<User>;
}
