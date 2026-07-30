import { artistsService } from '$lib/application/artists';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') as 'draft' | 'pending' | 'approved' | null;
  const artists = await requireAdmin(() => artistsService.list(status ?? undefined));
  return { artists, status: status ?? 'all' };
};
