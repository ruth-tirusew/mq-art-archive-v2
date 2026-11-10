import { truncateText } from '$lib/utils/answers';

/** The public address of an artist's short profile, e.g. https://artiv.et/@ada. */
export function profileUrl(origin: string, handle: string): string {
	return `${origin.replace(/\/$/, '')}/@${handle}`;
}

/**
 * The picture a link preview should show, or undefined. Social sites need a full http(s) address they can
 * fetch: a data: image (like the sample profile's swatches), a blob: preview or a relative path would show
 * nothing, or break the preview, so those are left out.
 */
export function shareImage(candidate: string | undefined | null): string | undefined {
	if (!candidate) return undefined;
	try {
		const url = new URL(candidate);
		return url.protocol === 'https:' || url.protocol === 'http:' ? url.href : undefined;
	} catch {
		return undefined;
	}
}

/** One or two sentences for a link preview: the tagline, else the start of the bio's words. */
export function shareDescription(tagline: string | undefined, bioText: string | undefined, max = 160): string {
	return (tagline?.trim() || truncateText(bioText ?? '', max)).trim();
}
