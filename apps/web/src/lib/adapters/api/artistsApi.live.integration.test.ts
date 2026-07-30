import { describe, expect, it } from 'vitest';
import { ArtistsApi } from '$lib/adapters/api/artistsApi';
import { liveEnabled, requireLiveApi } from '../../../test/live';

describe.skipIf(!liveEnabled())('ArtistsApi live @live', () => {
	it('lists artists from the running API', async () => {
		requireLiveApi();
		const api = new ArtistsApi();
		const artists = await api.list({ limit: 5 });
		expect(Array.isArray(artists)).toBe(true);
	});
});
