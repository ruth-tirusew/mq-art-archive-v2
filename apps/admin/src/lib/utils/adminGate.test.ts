import { describe, expect, it } from 'vitest';
import {
  ACCESS_DENIED_PATH,
  LOGIN_PATH,
  adminGateRedirect,
  adminOnlyRedirect,
  isAdminOnlyPath,
  isPublicPath,
  isStaffRole
} from '$lib/utils/adminGate';

describe('adminGateRedirect', () => {
  it('sends anonymous visitors to the login page', () => {
    expect(adminGateRedirect(null)).toBe(LOGIN_PATH);
  });
  it('explains the refusal to signed-in non-admins', () => {
    for (const role of ['public', 'artist', 'institution', 'contributor']) {
      expect(adminGateRedirect({ role })).toBe(ACCESS_DENIED_PATH);
    }
  });
  it('lets admins and moderators through', () => {
    expect(adminGateRedirect({ role: 'admin' })).toBeNull();
    expect(adminGateRedirect({ role: 'moderator' })).toBeNull();
  });
});

describe('isPublicPath', () => {
  it('leaves only login, accepting an invite and the OAuth callback open', () => {
    expect(isPublicPath('/login')).toBe(true);
    expect(isPublicPath('/accept-invite')).toBe(true);
    expect(isPublicPath('/auth/callback')).toBe(true);
    for (const path of ['/', '/users', '/loginx', '/authors', '/accept-invite/x', '/accept-invites']) expect(isPublicPath(path)).toBe(false);
  });
});

describe('staff roles and admin-only pages', () => {
  it('knows who is staff', () => {
    expect(isStaffRole('admin')).toBe(true);
    expect(isStaffRole('moderator')).toBe(true);
    for (const role of ['public', 'artist', 'institution', 'contributor', '']) expect(isStaffRole(role)).toBe(false);
  });

  it('keeps user management to admins, and nothing else', () => {
    expect(isAdminOnlyPath('/users')).toBe(true);
    expect(isAdminOnlyPath('/users/123')).toBe(true);
    for (const path of ['/', '/applications', '/settings', '/usersearch', '/posts']) expect(isAdminOnlyPath(path)).toBe(false);
  });

  it('sends a moderator away from admin-only pages, and nobody else', () => {
    expect(adminOnlyRedirect({ role: 'moderator' }, '/users')).toBe('/');
    expect(adminOnlyRedirect({ role: 'moderator' }, '/applications')).toBeNull();
    expect(adminOnlyRedirect({ role: 'admin' }, '/users')).toBeNull();
  });
});
