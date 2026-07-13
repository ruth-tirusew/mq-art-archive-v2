/** Split "Selamawit Abebe" into first / last for display typography. */
export function splitDisplayName(name: string): { first: string; last: string } {
	const parts = name.trim().split(/\s+/);
	if (parts.length === 1) return { first: parts[0], last: '' };
	return { first: parts[0], last: parts.slice(1).join(' ') };
}

export function postImageUrl(
	media?: { url: string }[],
	_postId?: string,
	_seed?: string
): string | undefined {
	return media?.[0]?.url;
}

export function formatEventDate(iso: string): string {
	return new Date(iso).toLocaleDateString('en-GB', {
		day: '2-digit',
		month: 'short',
		year: 'numeric'
	});
}

export function formatEventTime(iso: string): string {
	return new Date(iso).toLocaleTimeString('en-GB', {
		hour: '2-digit',
		minute: '2-digit',
		timeZoneName: 'short'
	});
}
