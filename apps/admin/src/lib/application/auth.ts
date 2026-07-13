import { writable } from 'svelte/store';
import { AuthApi } from '$lib/adapters/api/authApi';
import type { User } from '$lib/core/domain/auth';
import { googleLoginUrl } from '$lib/core/domain/auth';

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

  async logout(): Promise<void> {
    await api.logout();
    currentUser.set(null);
  }
};
