import { describe, expect, it, vi } from 'vitest';
import { createStaffInvitesService } from '$lib/application/staffInvites';
import { canResendStaffInvite, canRevokeStaffInvite, passwordProblem, staffInviteEmailProblem } from '$lib/core/domain/staffInvite';
import type { StaffInvitesPort } from '$lib/core/ports/staffInvites';

function fakePort(over: Partial<StaffInvitesPort> = {}): StaffInvitesPort {
  const invite = { id: 'i1', email: 'a@b.org', role: 'moderator' as const, status: 'live' as const, expires_at: '2026-12-01T00:00:00Z', created_at: '2026-11-10T00:00:00Z' };
  return {
    list: vi.fn().mockResolvedValue({ items: [], total: 0 }),
    create: vi.fn().mockResolvedValue({ invite, link: 'https://admin.test/accept-invite?token=t' }),
    resend: vi.fn().mockResolvedValue({ invite, link: 'https://admin.test/accept-invite?token=t2' }),
    revoke: vi.fn().mockResolvedValue(undefined),
    preview: vi.fn().mockResolvedValue({ email: 'a@b.org', role: 'moderator', status: 'live', expires_at: '2026-12-01T00:00:00Z' }),
    accept: vi.fn().mockResolvedValue({ id: 'u1', email: 'a@b.org', role: 'moderator', email_verified: true, has_password: true }),
    ...over
  };
}

describe('staff invites service', () => {
  it('refuses a bad address without calling the API, and trims a good one', async () => {
    const port = fakePort();
    const service = createStaffInvitesService(port);
    await expect(service.create('not-an-email')).rejects.toThrow('valid email');
    expect(port.create).not.toHaveBeenCalled();
    await service.create('  mod@example.org ');
    expect(port.create).toHaveBeenCalledWith('mod@example.org');
  });

  it('shows the API reason as it is (for example "already has an account")', async () => {
    const service = createStaffInvitesService(fakePort({ create: vi.fn().mockRejectedValue(new Error('a@b.org already has an Artiv account. Change their role on the Users page instead')) }));
    await expect(service.create('a@b.org')).rejects.toThrow('already has an Artiv account');
  });

  it('never sends a weak password: the one-time link is not spent on a typo', async () => {
    const port = fakePort();
    const service = createStaffInvitesService(port);
    await expect(service.accept('tok', 'short')).rejects.toThrow('at least 8');
    expect(port.accept).not.toHaveBeenCalled();
    await service.accept('tok', 'a-good-password');
    expect(port.accept).toHaveBeenCalledWith('tok', 'a-good-password');
  });
});

describe('staff invite rules', () => {
  it('counts password length in bytes, like the server, so Amharic is not accepted past the limit', () => {
    expect(passwordProblem('12345678')).toBeNull();
    expect(passwordProblem('ሰ'.repeat(24))).toBeNull(); // 24 letters = 72 bytes
    expect(passwordProblem('ሰ'.repeat(25))).toMatch(/too long/); // 75 bytes
    expect(passwordProblem('a'.repeat(73))).toMatch(/too long/);
  });

  it('checks addresses loosely (the server decides)', () => {
    expect(staffInviteEmailProblem('a@b.co')).toBeNull();
    for (const bad of ['', 'a', 'a@b', '@b.co', 'a b@c.de']) expect(staffInviteEmailProblem(bad)).not.toBeNull();
  });

  it('lets a waiting or expired invite be resent, but only a waiting one cancelled', () => {
    expect(canResendStaffInvite({ status: 'live' })).toBe(true);
    expect(canResendStaffInvite({ status: 'expired' })).toBe(true);
    expect(canResendStaffInvite({ status: 'revoked' })).toBe(true);
    expect(canResendStaffInvite({ status: 'claimed' })).toBe(false);
    expect(canRevokeStaffInvite({ status: 'live' })).toBe(true);
    for (const status of ['claimed', 'expired', 'revoked'] as const) expect(canRevokeStaffInvite({ status })).toBe(false);
  });
});
