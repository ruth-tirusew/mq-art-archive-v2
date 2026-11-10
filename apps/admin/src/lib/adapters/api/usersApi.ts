import { apiFetch, jsonRequest } from '$lib/adapters/api/client';
import type { Role, User } from '$lib/core/domain/user';
import type { UsersPort } from '$lib/core/ports/users';

export class UsersApi implements UsersPort {
  list({ limit, offset }: { limit: number; offset: number }): Promise<User[]> {
    return apiFetch<User[]>(`/admin/v1/users?limit=${limit}&offset=${offset}`);
  }

  changeRole(id: string, role: Role): Promise<User> {
    return apiFetch<User>(`/admin/v1/users/${encodeURIComponent(id)}`, jsonRequest('PATCH', { role }));
  }
}
