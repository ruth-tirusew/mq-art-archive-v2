import { writable } from 'svelte/store';
import { AuthApi } from '$lib/adapters/api/authApi';
import type { User } from '$lib/core/domain/auth';
import { googleLoginUrl } from '$lib/core/domain/auth';
import type {
  NotificationPreferences,
  UpdateProfileInput
} from '$lib/core/domain/settings';

const api = new AuthApi();

export const currentUser = writable<User | null>(null);
export const authLoading = writable(true);

export const authService = {
  googleLoginUrl,

  async load(): Promise<User | null> {
    authLoading.set(true);
    try {
      const user = await api.me();
      currentUser.set(user);
      return user;
    } catch {
      currentUser.set(null);
      return null;
    } finally {
      authLoading.set(false);
    }
  },

  async login(email: string, password: string): Promise<User> {
    const user = await api.login(email, password);
    currentUser.set(user);
    return user;
  },

  async register(email: string, password: string): Promise<User> {
    const user = await api.register(email, password);
    currentUser.set(user);
    return user;
  },

  async logout(): Promise<void> {
    await api.logout();
    currentUser.set(null);
  },

  async updateProfile(input: UpdateProfileInput): Promise<User> {
    const user = await api.updateProfile(input);
    currentUser.set(user);
    return user;
  },

  async changeEmail(email: string, currentPassword: string): Promise<User> {
    const user = await api.changeEmail(email, currentPassword);
    currentUser.set(user);
    return user;
  },

  async changePassword(currentPassword: string, newPassword: string): Promise<void> {
    await api.changePassword(currentPassword, newPassword);
  },

  getNotifications(): Promise<NotificationPreferences> {
    return api.getNotifications();
  },

  updateNotifications(prefs: NotificationPreferences): Promise<NotificationPreferences> {
    return api.updateNotifications(prefs);
  }
};
