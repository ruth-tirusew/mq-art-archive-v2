import { apiFetch, apiFetchPage, jsonRequest } from '$lib/adapters/api/client';
import type { StaffInvite, StaffInviteCreated, StaffInvitePage, StaffInvitePreview } from '$lib/core/domain/staffInvite';
import type { User } from '$lib/core/domain/auth';
import type { StaffInvitesPort } from '$lib/core/ports/staffInvites';

export class StaffInvitesApi implements StaffInvitesPort {
  list({ limit, offset }: { limit: number; offset: number }): Promise<StaffInvitePage> {
    return apiFetchPage<StaffInvite>(`/admin/v1/staff-invites?limit=${limit}&offset=${offset}`);
  }

  create(email: string): Promise<StaffInviteCreated> {
    return apiFetch<StaffInviteCreated>('/admin/v1/staff-invites', jsonRequest('POST', { email, role: 'moderator' }));
  }

  resend(id: string): Promise<StaffInviteCreated> {
    return apiFetch<StaffInviteCreated>(`/admin/v1/staff-invites/${encodeURIComponent(id)}/resend`, jsonRequest('POST'));
  }

  revoke(id: string): Promise<void> {
    return apiFetch<void>(`/admin/v1/staff-invites/${encodeURIComponent(id)}`, { method: 'DELETE' });
  }

  preview(token: string): Promise<StaffInvitePreview> {
    return apiFetch<StaffInvitePreview>('/api/v1/staff-invites/preview', jsonRequest('POST', { token }));
  }

  accept(token: string, password: string): Promise<User> {
    return apiFetch<User>('/api/v1/staff-invites/accept', jsonRequest('POST', { token, password }));
  }
}
