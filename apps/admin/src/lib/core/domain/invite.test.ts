import { describe, expect, it } from 'vitest';
import {
  canResendInvite,
  canRevokeInvite,
  normalizeInviteHandle,
  parseInviteStatusFilter,
  validateInviteForm
} from '$lib/core/domain/invite';
import { reviewBlocker } from '$lib/core/domain/onboarding';

describe('validateInviteForm', () => {
  it('requires a plausible email', () => {
    expect(validateInviteForm({ email: 'ada@example.com', handle: '' })).toEqual({});
    expect(validateInviteForm({ email: 'nope', handle: '' }).email).toBeDefined();
  });
  it('treats the handle as optional and case-insensitive, like the API', () => {
    expect(validateInviteForm({ email: 'a@b.co', handle: '' })).toEqual({});
    expect(validateInviteForm({ email: 'a@b.co', handle: ' Ada_1 ' })).toEqual({});
    expect(normalizeInviteHandle(' Ada_1 ')).toBe('ada_1');
    expect(validateInviteForm({ email: 'a@b.co', handle: 'a-b' }).handle).toBeDefined();
    expect(validateInviteForm({ email: 'a@b.co', handle: 'ab' }).handle).toBeDefined();
  });
});

describe('invite actions', () => {
  it('allows resending anything but a used invite, and revoking only a waiting one', () => {
    expect(['live', 'expired', 'revoked'].every((status) => canResendInvite({ status: status as never }))).toBe(true);
    expect(canResendInvite({ status: 'claimed' })).toBe(false);
    expect(canRevokeInvite({ status: 'live' })).toBe(true);
    for (const status of ['claimed', 'expired', 'revoked'] as const) expect(canRevokeInvite({ status })).toBe(false);
  });
});

describe('reviewBlocker', () => {
  it('needs a reason to reject but not to approve', () => {
    expect(reviewBlocker('rejected', '  ')).not.toBeNull();
    expect(reviewBlocker('rejected', 'Please add links')).toBeNull();
    expect(reviewBlocker('approved', '')).toBeNull();
  });
});

describe('parseInviteStatusFilter', () => {
  it('accepts the API statuses and falls back to everything for anything else', () => {
    expect(parseInviteStatusFilter('expired')).toBe('expired');
    expect(parseInviteStatusFilter('all')).toBe('all');
    expect(parseInviteStatusFilter(null)).toBe('all');
    expect(parseInviteStatusFilter('bogus')).toBe('all');
  });
});
