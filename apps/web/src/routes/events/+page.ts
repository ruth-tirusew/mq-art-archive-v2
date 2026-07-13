import { eventsService } from '$lib/application/events';
import { loadWithApiFallback } from '$lib/config/loadApi';
import { events } from '$lib/data/events';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const type = url.searchParams.get('type') ?? undefined;

	return loadWithApiFallback(
		async () => {
			const apiEvents = await eventsService.list({ type, limit: 50 });
			const types = [...new Set(apiEvents.map((e) => e.event_type).filter(Boolean))].sort();
			return { source: 'api' as const, events: apiEvents, types, filterType: type };
		},
		() => {
			const types = [...new Set(events.map((e) => e.type))].sort();
			return { source: 'static' as const, events, types, filterType: undefined };
		}
	);
};
