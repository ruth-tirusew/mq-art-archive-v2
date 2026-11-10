import {
  STAFF_INVITES_PAGE_SIZE,
  passwordProblem,
  staffInviteEmailProblem,
  type StaffInviteCreated,
  type StaffInvitePage,
  type StaffInvitePreview
} from '$lib/core/domain/staffInvite';
import type { User } from '$lib/core/domain/auth';
import type { StaffInvitesPort } from '$lib/core/ports/staffInvites';

export function createStaffInvitesService(port: StaffInvitesPort) {
  return {
    list(offset = 0): Promise<StaffInvitePage> {
      return port.list({ limit: STAFF_INVITES_PAGE_SIZE, offset });
    },
    /** Checks the address before bothering the server; the server's answer (e.g. "already has an account") is shown as is. */
    create(email: string): Promise<StaffInviteCreated> {
      const problem = staffInviteEmailProblem(email);
      if (problem) return Promise.reject(new Error(problem));
      return port.create(email.trim());
    },
    resend(id: string): Promise<StaffInviteCreated> {
      return port.resend(id);
    },
    revoke(id: string): Promise<void> {
      return port.revoke(id);
    },
    preview(token: string): Promise<StaffInvitePreview> {
      return port.preview(token);
    },
    /** A weak password is refused here, so a typo never reaches the one-time link. */
    accept(token: string, password: string): Promise<User> {
      const problem = passwordProblem(password);
      if (problem) return Promise.reject(new Error(problem));
      return port.accept(token, password);
    }
  };
}
