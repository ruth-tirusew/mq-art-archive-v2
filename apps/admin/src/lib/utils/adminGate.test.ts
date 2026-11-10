import { describe, expect, it } from 'vitest';
import { ACCESS_DENIED_PATH, LOGIN_PATH, adminGateRedirect, isPublicPath } from '$lib/utils/adminGate';

describe('adminGateRedirect', () => {
  it('sends anonymous visitors to the login page', () => {
    expect(adminGateRedirect(null)).toBe(LOGIN_PATH);
  });
  it('explains the refusal to signed-in non-admins', () => {
    for (const role of ['public', 'artist', 'institution', 'contributor']) {
      expect(adminGateRedirect({ role })).toBe(ACCESS_DENIED_PATH);
    }
  });
  it('lets admins through', () => {
    expect(adminGateRedirect({ role: 'admin' })).toBeNull();
  });
});

describe('isPublicPath', () => {
  it('leaves only login and the OAuth callback open', () => {
    expect(isPublicPath('/login')).toBe(true);
    expect(isPublicPath('/auth/callback')).toBe(true);
    for (const path of ['/', '/users', '/loginx', '/authors']) expect(isPublicPath(path)).toBe(false);
  });
});
