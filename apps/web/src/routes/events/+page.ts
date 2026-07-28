import { eventsService } from '$lib/application/events';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const type = url.searchParams.get('type') ?? undefined;

	try {
		// Load full calendar; type/upcoming filters are applied client-side.
		const events = await eventsService.list({ upcoming: false, limit: 100 });
		const types = [...new Set(events.map((event) => event.event_type).filter(Boolean))].sort();
		return { error: false as const, events, types, filterType: type };
	} catch {
		return { error: true as const, events: [], types: [], filterType: type };
	}
};
