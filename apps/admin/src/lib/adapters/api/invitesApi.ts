import { apiFetch, jsonRequest } from '$lib/adapters/api/client';
import type { CreateInviteInput, Invite, InviteCreated } from '$lib/core/domain/invite';
import type { InvitesPort } from '$lib/core/ports/invites';

export class InvitesApi implements InvitesPort {
  list(): Promise<Invite[]> {
    return apiFetch<Invite[]>('/admin/v1/invites');
  }

  create(input: CreateInviteInput): Promise<InviteCreated> {
    return apiFetch<InviteCreated>('/admin/v1/invites', jsonRequest('POST', input));
  }

  resend(id: string): Promise<InviteCreated> {
    return apiFetch<InviteCreated>(`/admin/v1/invites/${encodeURIComponent(id)}/resend`, jsonRequest('POST'));
  }

  revoke(id: string): Promise<void> {
    return apiFetch<void>(`/admin/v1/invites/${encodeURIComponent(id)}`, { method: 'DELETE' });
  }
}
