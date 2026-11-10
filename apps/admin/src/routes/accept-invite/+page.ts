import { ApiError } from '$lib/adapters/api/client';
import { staffInvitesService } from '$lib/composition';
import type { StaffInvitePreview } from '$lib/core/domain/staffInvite';
import type { PageLoad } from './$types';

export type AcceptState =
  | { kind: 'ready'; preview: StaffInvitePreview }
  | { kind: 'used' | 'revoked' | 'expired'; preview: StaffInvitePreview }
  /** No token, or one that matches no invite. Deliberately the same message: it says nothing about which. */
  | { kind: 'invalid' }
  | { kind: 'error' };

// Public page: the invitee has no account yet, so there is no sign-in gate here (see isPublicPath). The
// token is in the address only because it arrives in an email link; it is sent to the API in a POST body.
export const load: PageLoad = async ({ url }) => {
  const token = url.searchParams.get('token')?.trim() ?? '';
  let state: AcceptState;
  if (!token) {
    state = { kind: 'invalid' };
  } else {
    try {
      const preview = await staffInvitesService.preview(token);
      const kind = { live: 'ready', claimed: 'used', revoked: 'revoked', expired: 'expired' } as const;
      state = { kind: kind[preview.status], preview } as AcceptState;
    } catch (cause) {
      state = cause instanceof ApiError && cause.status === 404 ? { kind: 'invalid' } : { kind: 'error' };
    }
  }
  return { token, state };
};
