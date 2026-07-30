import { PUBLIC_API_URL } from '$env/static/public';

/** Skip unless `LIVE_API=1` is set (Makefile `web-test-live`). */
export function requireLiveApi(): void {
	if (process.env.LIVE_API !== '1') {
		throw new Error('LIVE_API must be 1 to run live integration tests');
	}
	if (!PUBLIC_API_URL) {
		throw new Error('PUBLIC_API_URL must be set for live integration tests');
	}
}

export function liveEnabled(): boolean {
	return process.env.LIVE_API === '1' && Boolean(PUBLIC_API_URL);
}
