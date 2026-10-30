import { artistsService } from '$lib/application/artists';
import { requireAdmin } from '$lib/utils/loadGuard';
import { PAGE_SIZE } from '$lib/utils/paging';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') as 'draft' | 'pending' | 'approved' | null;
  const artists = await requireAdmin(() => artistsService.list(status ?? undefined, { limit: PAGE_SIZE }));
  return { artists, status: status ?? 'all' };
};
