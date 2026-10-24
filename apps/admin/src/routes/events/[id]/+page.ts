import { eventsService } from '$lib/application/events';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
  requireAdmin(async () => ({ event: await eventsService.getById(params.id) }));
