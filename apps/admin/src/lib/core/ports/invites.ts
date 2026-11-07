import type { CreateInviteInput, InviteCreated, InvitePage, InviteStatusFilter } from '$lib/core/domain/invite';

export interface InvitesPort {
  list(status: InviteStatusFilter, page: { limit: number; offset: number }): Promise<InvitePage>;
  create(input: CreateInviteInput): Promise<InviteCreated>;
  resend(id: string): Promise<InviteCreated>;
  revoke(id: string): Promise<void>;
}
