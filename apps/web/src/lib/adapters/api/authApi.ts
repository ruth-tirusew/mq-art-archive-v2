import { apiFetch } from '$lib/adapters/api/client';
import type { User } from '$lib/core/domain/auth';
import type { AuthPort } from '$lib/core/ports/auth';

export class AuthApi implements AuthPort {
  me(): Promise<User> {
    return apiFetch<User>('/api/v1/auth/me');
  }

  login(email: string, password: string): Promise<User> {
    return apiFetch<User>('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
  }

  register(email: string, password: string): Promise<User> {
    return apiFetch<User>('/api/v1/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
  }

  logout(): Promise<void> {
    return apiFetch<void>('/api/v1/auth/logout', { method: 'POST' });
  }

  forgotPassword(email: string): Promise<void> {
    return apiFetch<void>('/api/v1/auth/forgot-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email })
    });
  }

  resetPassword(token: string, password: string): Promise<void> {
    return apiFetch<void>('/api/v1/auth/reset-password', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ token, password })
    });
  }
}
