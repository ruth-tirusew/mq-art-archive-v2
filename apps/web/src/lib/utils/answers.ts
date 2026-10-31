import type { PublicAnswer } from '$lib/core/domain/questionnaire';

/**
 * Returns a normalised URL only for plain http(s) links with a host and no embedded credentials,
 * otherwise null. Link answers come from users, so this is the gate before anything becomes an
 * href: javascript:, data:, vbscript:, file:, protocol-relative and bare-host values all fail.
 */
export function safeHttpUrl(value: string): string | null {
	const raw = value.trim();
	if (!raw) return null;
	let url: URL;
	try {
		url = new URL(raw);
	} catch {
		return null;
	}
	if (url.protocol !== 'http:' && url.protocol !== 'https:') return null;
	if (!url.hostname || url.username || url.password) return null;
	return url.href;
}

/** A short, readable form of a URL for link text: host and path without scheme or trailing slash. */
export function linkLabel(href: string): string {
	try {
		const url = new URL(href);
		return `${url.host}${url.pathname === '/' ? '' : url.pathname}`.replace(/\/$/, '');
	} catch {
		return href;
	}
}

/** The text of a plain or choice answer. A choice is stored as the option's value; show its label. */
export function answerText(answer: PublicAnswer): string {
	if (answer.type === 'choice') {
		return answer.options?.find((o) => o.value === answer.body)?.label_en ?? answer.body;
	}
	return answer.body;
}
