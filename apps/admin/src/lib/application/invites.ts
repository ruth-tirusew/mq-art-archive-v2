import {
  INVITES_PAGE_SIZE,
  type CreateInviteInput,
  type InviteCreated,
  type InvitePage,
  type InviteStatusFilter
} from '$lib/core/domain/invite';
import type { InvitesPort } from '$lib/core/ports/invites';

export function createInvitesService(port: InvitesPort) {
  return {
    list(status: InviteStatusFilter, offset = 0): Promise<InvitePage> {
      return port.list(status, { limit: INVITES_PAGE_SIZE, offset });
    },
    create(input: CreateInviteInput): Promise<InviteCreated> {
      return port.create(input);
    },
    resend(id: string): Promise<InviteCreated> {
      return port.resend(id);
    },
    revoke(id: string): Promise<void> {
      return port.revoke(id);
    }
  };
}
