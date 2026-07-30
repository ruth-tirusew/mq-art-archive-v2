import { describe, expect, it } from 'vitest';
import { searchAll } from './search';

describe('searchAll', () => {
	it('maps the unified search response', async () => {
		const results = await searchAll('ceramics', 4);
		expect(results.artists[0]?.display_name).toBe('Selamawit Tesfaye');
		expect(results.posts[0]?.title).toBe('Blue Market');
		expect(results.articles[0]?.slug).toBe('ceramics');
		expect(results.events[0]?.slug).toBe('opening');
	});

	it('does not request an empty query', async () => {
		await expect(searchAll('  ')).resolves.toEqual({
			artists: [],
			posts: [],
			articles: [],
			events: []
		});
	});
});
