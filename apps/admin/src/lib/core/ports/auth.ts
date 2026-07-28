import type { User } from '$lib/core/domain/auth';
import type {
  NotificationPreferences,
  UpdateProfileInput
} from '$lib/core/domain/settings';

export interface AuthPort {
  me(): Promise<User>;
  login(email: string, password: string): Promise<User>;
  register(email: string, password: string): Promise<User>;
  logout(): Promise<void>;
  updateProfile(input: UpdateProfileInput): Promise<User>;
  changeEmail(email: string, currentPassword: string): Promise<User>;
  changePassword(currentPassword: string, newPassword: string): Promise<void>;
  getNotifications(): Promise<NotificationPreferences>;
  updateNotifications(prefs: NotificationPreferences): Promise<NotificationPreferences>;
}
