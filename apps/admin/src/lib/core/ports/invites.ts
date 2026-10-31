import type { CreateInviteInput, Invite, InviteCreated } from '$lib/core/domain/invite';

export interface InvitesPort {
  list(): Promise<Invite[]>;
  create(input: CreateInviteInput): Promise<InviteCreated>;
  resend(id: string): Promise<InviteCreated>;
  revoke(id: string): Promise<void>;
}
