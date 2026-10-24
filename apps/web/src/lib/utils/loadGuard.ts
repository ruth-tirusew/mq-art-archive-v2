import { error, isHttpError, isRedirect } from '@sveltejs/kit';
import { ApiError } from '$lib/adapters/api/client';

/**
 * Run a public loader so failures reach +error.svelte with the right status instead of being
 * swallowed: a missing record is a real 404, the API being down or erroring is a retryable 503.
 * Without this, every failure looked like "not found" or an empty page.
 */
export async function guardLoad<T>(load: () => Promise<T>, notFoundMessage = 'We could not find that page.'): Promise<T> {
	try {
		return await load();
	} catch (err) {
		if (isHttpError(err) || isRedirect(err)) throw err;
		if (err instanceof ApiError) {
			if (err.status === 404) error(404, notFoundMessage);
			if (err.status === 401 || err.status === 403) error(err.status, err.message);
			if (err.status < 500) error(err.status, err.message);
		}
		error(503, 'We could not reach the archive.');
	}
}
