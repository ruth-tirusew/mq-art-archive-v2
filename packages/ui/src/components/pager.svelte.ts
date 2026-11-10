export type PagerOptions<T> = {
	/** Rows already fetched by the page's loader (the first page). */
	initial: T[];
	pageSize: number;
	/** Fetches the next slice starting at `offset`. */
	fetchPage: (offset: number, limit: number) => Promise<T[]>;
};

/**
 * "Load more" for a list. The route loader fetches the first page, the pager appends the rest on
 * demand. A full page back means there may be more; a short page means everything is loaded.
 *   const pager = createPager({ initial: data.artists, pageSize: 50, fetchPage });
 *   $effect(() => pager.reset(data.artists));   // when the loader's data changes (a filter)
 */
export function createPager<T>(options: PagerOptions<T>) {
	let items = $state<T[]>(options.initial);
	let hasMore = $state(options.initial.length >= options.pageSize);
	let loading = $state(false);
	let error = $state('');

	return {
		get items() {
			return items;
		},
		get hasMore() {
			return hasMore;
		},
		get loading() {
			return loading;
		},
		get error() {
			return error;
		},
		reset(initial: T[]) {
			items = initial;
			hasMore = initial.length >= options.pageSize;
			error = '';
		},
		async loadMore() {
			if (loading || !hasMore) return;
			loading = true;
			error = '';
			try {
				const next = await options.fetchPage(items.length, options.pageSize);
				items = [...items, ...next];
				hasMore = next.length >= options.pageSize;
			} catch (e) {
				error = e instanceof Error ? e.message : 'Could not load more';
			} finally {
				loading = false;
			}
		}
	};
}
