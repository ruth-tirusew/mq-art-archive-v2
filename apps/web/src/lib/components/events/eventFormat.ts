import type { Event } from '$lib/core/domain/events';

export function fmtEventDate(iso: string) {
	const d = new Date(iso);
	return {
		day: d.getDate().toString().padStart(2, '0'),
		month: d.toLocaleString('en', { month: 'short' }).toUpperCase(),
		weekday: d.toLocaleString('en', { weekday: 'short' }).toUpperCase(),
		weekdayLong: d.toLocaleString('en', { weekday: 'long' }).toUpperCase(),
		time: d.toLocaleTimeString('en', { hour: 'numeric', minute: '2-digit' }),
		dateKey: `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`,
		full: d.toLocaleDateString('en-GB', {
			weekday: 'long',
			day: 'numeric',
			month: 'long',
			year: 'numeric'
		})
	};
}

export function truncate(text: string | undefined, max = 120): string {
	if (!text) return '';
	const t = text.trim().replace(/\s+/g, ' ');
	if (t.length <= max) return t;
	return t.slice(0, max - 1).trimEnd() + '…';
}

export function startOfWeek(d = new Date()): Date {
	const x = new Date(d);
	const day = x.getDay();
	const diff = day === 0 ? -6 : 1 - day;
	x.setHours(0, 0, 0, 0);
	x.setDate(x.getDate() + diff);
	return x;
}

export function endOfWeek(d = new Date()): Date {
	const start = startOfWeek(d);
	const end = new Date(start);
	end.setDate(end.getDate() + 7);
	return end;
}

export function isSameDay(a: Date, b: Date): boolean {
	return (
		a.getFullYear() === b.getFullYear() &&
		a.getMonth() === b.getMonth() &&
		a.getDate() === b.getDate()
	);
}

export function weekStats(events: Event[]) {
	const start = startOfWeek();
	const end = endOfWeek();
	const inWeek = events.filter((e) => {
		const t = new Date(e.starts_at).getTime();
		return t >= start.getTime() && t < end.getTime();
	});

	const byType = (needle: string) =>
		inWeek.filter((e) => e.event_type.toLowerCase().includes(needle)).length;

	return {
		total: inWeek.length,
		openings: byType('opening'),
		workshops: byType('workshop'),
		exhibitions: byType('exhibition'),
		talks: byType('talk') + byType('reading') + byType('poetry')
	};
}

export function groupEventsByDay(events: Event[]): { key: string; label: string; events: Event[] }[] {
	const map = new Map<string, Event[]>();
	for (const e of events) {
		const d = fmtEventDate(e.starts_at);
		const list = map.get(d.dateKey) ?? [];
		list.push(e);
		map.set(d.dateKey, list);
	}
	return [...map.entries()]
		.sort(([a], [b]) => a.localeCompare(b))
		.map(([key, items]) => {
			const d = fmtEventDate(items[0].starts_at);
			return {
				key,
				label: `${d.weekdayLong}, ${d.month} ${Number(d.day)}`,
				events: items.sort(
					(a, b) => new Date(a.starts_at).getTime() - new Date(b.starts_at).getTime()
				)
			};
		});
}
