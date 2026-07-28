import { PUBLIC_API_URL } from '$env/static/public';

export function getApiBaseUrl(): string {
	const baseUrl = PUBLIC_API_URL.trim().replace(/\/$/, '');
	if (!baseUrl) {
		throw new Error('PUBLIC_API_URL is required to load Artiv product data');
	}
	return baseUrl;
}
