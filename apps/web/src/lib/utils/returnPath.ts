/**
 * Accept only same-site relative paths ("/studio", "/wiki/x?y=1"). Anything else — absolute URLs,
 * protocol-relative "//evil.com", backslash tricks, control characters — falls back, so a crafted
 * `?return_to=` can never send a signed-in user to another site.
 */
export function safeReturnPath(raw: string | null | undefined, fallback = '/studio'): string {
	if (!raw) return fallback;
	if (!raw.startsWith('/') || raw.startsWith('//') || raw.startsWith('/\\')) return fallback;
	if (/[\u0000-\u001f\\]/.test(raw)) return fallback;
	return raw;
}
