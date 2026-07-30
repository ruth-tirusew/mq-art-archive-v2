import { describe, expect, it } from 'vitest';
import { apiFetch } from '$lib/adapters/api/client';
import { liveEnabled, requireLiveApi } from '../../../test/live';

/** Seeded admin from `00009_seed_dev_admin.sql` (requires AUTH_DEV_MODE). */
const DEV_ADMIN_ID =
	process.env.LIVE_USER_ID ?? '00000000-0000-4000-8000-000000000001';

describe.skipIf(!liveEnabled())('Admin posts live @live', () => {
	it('lists posts from the running API', async () => {
		requireLiveApi();
		const posts = await apiFetch<unknown[]>('/admin/v1/posts', {
			headers: { 'X-User-ID': DEV_ADMIN_ID }
		});
		expect(Array.isArray(posts)).toBe(true);
	});
});
