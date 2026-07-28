import { eventsService } from '$lib/application/events';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		const event = await eventsService.getBySlug(params.slug);
		return { error: false as const, event };
	} catch {
		return { error: true as const, event: null };
	}
};
