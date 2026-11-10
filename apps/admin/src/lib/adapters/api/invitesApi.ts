import { apiFetch, apiFetchPage, jsonRequest } from '$lib/adapters/api/client';
import type { CreateInviteInput, Invite, InviteCreated, InvitePage, InviteStatusFilter } from '$lib/core/domain/invite';
import type { InvitesPort } from '$lib/core/ports/invites';

export class InvitesApi implements InvitesPort {
  list(status: InviteStatusFilter, { limit, offset }: { limit: number; offset: number }): Promise<InvitePage> {
    return apiFetchPage<Invite>(`/admin/v1/invites?status=${encodeURIComponent(status)}&limit=${limit}&offset=${offset}`);
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
