export type Role = 'public' | 'artist' | 'institution' | 'contributor' | 'admin';

export const ROLES: readonly Role[] = ['public', 'artist', 'institution', 'contributor', 'admin'];

export interface User {
  id: string;
  email: string;
  role: Role;
  display_name?: string;
}

/** Users are listed this many at a time; a full page means there may be more. */
export const USERS_PAGE_SIZE = 200;
