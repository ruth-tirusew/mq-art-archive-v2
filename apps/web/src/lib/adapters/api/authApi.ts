import { apiFetch } from '$lib/adapters/api/client';
import type { User } from '$lib/core/domain/auth';
import type { AuthPort } from '$lib/core/ports/auth';

export class AuthApi implements AuthPort {
  me(): Promise<User> {
    return apiFetch<User>('/api/v1/auth/me');
  }

  logout(): Promise<void> {
    return apiFetch<void>('/api/v1/auth/logout', { method: 'POST' });
  }
}
