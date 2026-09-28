import type { NotificationPreferences, User } from '$lib/core/domain/auth';

export interface AuthPort {
  me(): Promise<User>;
  login(email: string, password: string): Promise<User>;
  register(email: string, password: string): Promise<User>;
  logout(): Promise<void>;
  forgotPassword(email: string): Promise<void>;
  resetPassword(token: string, password: string): Promise<void>;
  updateProfile(displayName: string, avatarUrl: string): Promise<User>;
  changeEmail(email: string, currentPassword: string): Promise<User>;
  changePassword(currentPassword: string, newPassword: string): Promise<void>;
  getNotifications(): Promise<NotificationPreferences>;
  updateNotifications(prefs: NotificationPreferences): Promise<NotificationPreferences>;
}
