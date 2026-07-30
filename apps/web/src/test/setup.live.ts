import '@testing-library/jest-dom/vitest';
import { beforeAll } from 'vitest';

beforeAll(() => {
	if (process.env.LIVE_API !== '1') {
		console.warn(
			'[live] Skipping live API tests. Run with LIVE_API=1 and a local API on PUBLIC_API_URL.'
		);
	}
});
