import { eventsService } from '$lib/application/events';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') ?? 'pending';
  const events = await requireAdmin(() =>
    eventsService.list(status as 'all' | 'pending' | 'approved' | 'rejected')
  );
  return { events, status };
};
