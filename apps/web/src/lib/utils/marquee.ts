import type { Event as ApiEvent } from '$lib/core/domain/events';

function placeLabel(venue?: string, city?: string): string | undefined {
	const parts = [venue, city].filter(Boolean);
	return parts.length > 0 ? parts.join(', ') : undefined;
}

export function marqueeLabelFromEvent(event: {
	title: string;
	venue?: string;
	city?: string;
}): string {
	const place = placeLabel(event.venue, event.city);
	return place ? `${event.title} · ${place}` : event.title;
}

export function marqueeItemsFromApiEvents(events: ApiEvent[]): string[] {
	return events.map((event) => marqueeLabelFromEvent(event)).filter(Boolean);
}
