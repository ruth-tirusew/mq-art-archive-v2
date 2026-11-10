import { usersService } from '$lib/composition';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = () => requireAdmin(async () => ({ users: await usersService.listPage() }));
