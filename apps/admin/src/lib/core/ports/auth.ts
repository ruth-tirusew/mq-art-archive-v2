import type { User } from '$lib/core/domain/auth';

export interface AuthPort {
  me(): Promise<User>;
  logout(): Promise<void>;
}
