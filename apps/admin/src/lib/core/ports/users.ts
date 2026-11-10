import type { Role, User } from '$lib/core/domain/user';

export interface UsersPort {
  list(options: { limit: number; offset: number }): Promise<User[]>;
  changeRole(id: string, role: Role): Promise<User>;
}
