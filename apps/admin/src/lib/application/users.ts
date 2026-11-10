import { USERS_PAGE_SIZE, type Role, type User } from '$lib/core/domain/user';
import type { UsersPort } from '$lib/core/ports/users';

export function createUsersService(port: UsersPort) {
  return {
    /** One page of users, starting after `offset` already-loaded rows. */
    listPage(offset = 0): Promise<User[]> {
      return port.list({ limit: USERS_PAGE_SIZE, offset });
    },
    changeRole(id: string, role: Role): Promise<User> {
      return port.changeRole(id, role);
    }
  };
}
