import { describe, expect, it } from 'vitest';
import { AuthApi } from '$lib/adapters/api/authApi';

describe('AuthApi.verifyEmail (MSW)', () => {
	const api = new AuthApi();

	it('resolves for a valid token', async () => {
		await expect(api.verifyEmail('good')).resolves.toBeUndefined();
	});

	it('rejects an invalid or expired token with an invalid-kind error', async () => {
		await expect(api.verifyEmail('stale')).rejects.toMatchObject({ status: 400, kind: 'invalid' });
	});
});
