import type { CreateInviteInput, Invite, InviteCreated } from '$lib/core/domain/invite';
import type { InvitesPort } from '$lib/core/ports/invites';

export function createInvitesService(port: InvitesPort) {
  return {
    list(): Promise<Invite[]> {
      return port.list();
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
