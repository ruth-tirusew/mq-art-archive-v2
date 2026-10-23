import { artistsService } from '$lib/application/artists';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = () => requireAdmin(async () => ({ artists: await artistsService.list() }));
