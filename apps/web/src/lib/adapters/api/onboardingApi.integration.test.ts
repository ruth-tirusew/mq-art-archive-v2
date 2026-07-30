import { describe, expect, it } from 'vitest';
import { OnboardingApi } from './onboardingApi';

describe('OnboardingApi', () => {
	const api = new OnboardingApi();

	it('checks handle availability', async () => {
		await expect(api.checkHandleAvailability('new_artist')).resolves.toEqual({
			handle: 'new_artist',
			available: true
		});
		await expect(api.checkHandleAvailability('taken')).resolves.toMatchObject({ available: false });
	});
});
