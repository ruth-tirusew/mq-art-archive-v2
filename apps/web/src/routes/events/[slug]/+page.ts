import { eventsService } from '$lib/application/events';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => ({ event: await eventsService.getBySlug(params.slug) }), 'We could not find that event.');
