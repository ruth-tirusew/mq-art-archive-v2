import { eventsService } from '$lib/application/events';
import { useApi } from '$lib/config/dataSource';
import { getEvent } from '$lib/data/events';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	if (!useApi) {
		const event = getEvent(params.slug);
		if (!event) error(404, 'Event not found');
		return { source: 'static' as const, event };
	}

	try {
		const event = await eventsService.getBySlug(params.slug);
		return { source: 'api' as const, event };
	} catch {
		const event = getEvent(params.slug);
		if (event) return { source: 'static' as const, event };
		error(404, 'Event not found');
	}
};
