/** Vitest stand-in for `$env/static/public`. */
export const PUBLIC_API_URL =
	process.env.PUBLIC_API_URL ?? 'http://localhost:8080';
