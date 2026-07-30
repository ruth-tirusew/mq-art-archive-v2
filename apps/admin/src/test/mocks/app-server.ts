/** Vitest stand-in for `$app/server` — no request context in unit/integration tests. */
export function getRequestEvent(): never {
	throw new Error('No request event outside SvelteKit');
}
