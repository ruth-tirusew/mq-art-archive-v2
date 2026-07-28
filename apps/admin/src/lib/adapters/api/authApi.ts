import { apiFetch } from '$lib/adapters/api/client';
import type { User } from '$lib/core/domain/auth';
import type {
  NotificationPreferences,
  UpdateProfileInput
} from '$lib/core/domain/settings';
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

  updateProfile(input: UpdateProfileInput): Promise<User> {
    return apiFetch<User>('/api/v1/auth/me/profile', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(input)
    });
  }

  changeEmail(email: string, currentPassword: string): Promise<User> {
    return apiFetch<User>('/api/v1/auth/me/email', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, current_password: currentPassword })
    });
  }

  changePassword(currentPassword: string, newPassword: string): Promise<void> {
    return apiFetch<void>('/api/v1/auth/me/password', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        current_password: currentPassword,
        new_password: newPassword
      })
    });
  }

  getNotifications(): Promise<NotificationPreferences> {
    return apiFetch<NotificationPreferences>('/api/v1/auth/me/notifications');
  }

  updateNotifications(prefs: NotificationPreferences): Promise<NotificationPreferences> {
    return apiFetch<NotificationPreferences>('/api/v1/auth/me/notifications', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(prefs)
    });
  }
}
