<script lang="ts" module>
	import { getContext, setContext } from 'svelte';
	import { goto } from '$app/navigation';

	const SEARCH_KEY = Symbol('search');

	export function isMacPlatform() {
		if (typeof navigator === 'undefined') return false;
		return /Mac|iPhone|iPod|iPad/i.test(navigator.platform);
	}

	export function searchShortcutLabel() {
		return isMacPlatform() ? '⌘K' : 'Ctrl+K';
	}

	export function setSearchContext(value: { open: boolean; setOpen: (v: boolean) => void }) {
		setContext(SEARCH_KEY, value);
	}

	export function getSearchContext() {
		return getContext<{ open: boolean; setOpen: (v: boolean) => void }>(SEARCH_KEY);
	}

	type SearchResult = {
		label: string;
		meta: string;
		href: string;
		group: string;
		keywords: string;
	};

	export function buildSearchIndex(): SearchResult[] {
		return [];
	}
</script>

<script lang="ts">
	import { searchAll, type SearchResults } from '$lib/application/search';
	import { useApi } from '$lib/config/dataSource';

	let { children }: { children: import('svelte').Snippet } = $props();

	let open = $state(false);
	const setOpen = (v: boolean) => {
		open = v;
	};

	setSearchContext({ get open() { return open; }, setOpen });

	let query = $state('');
	let loading = $state(false);
	let apiResults = $state<SearchResults>({ artists: [], posts: [], articles: [], events: [] });
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;

	const staticIndex = buildSearchIndex();

	const staticFiltered = $derived.by(() => {
		const q = query.trim().toLowerCase();
		if (!q) return staticIndex;
		return staticIndex.filter((r) => r.keywords.toLowerCase().includes(q));
	});

	const staticGrouped = $derived.by(() => {
		const map = new Map<string, typeof staticFiltered>();
		for (const r of staticFiltered) {
			const list = map.get(r.group) ?? [];
			list.push(r);
			map.set(r.group, list);
		}
		return map;
	});

	const apiHasResults = $derived(
		apiResults.artists.length +
			apiResults.posts.length +
			apiResults.articles.length +
			apiResults.events.length >
			0
	);

	$effect(() => {
		if (!useApi) return;
		const q = query.trim();
		clearTimeout(debounceTimer);
		if (!q) {
			apiResults = { artists: [], posts: [], articles: [], events: [] };
			loading = false;
			return;
		}
		debounceTimer = setTimeout(async () => {
			loading = true;
			try {
				apiResults = await searchAll(q);
			} finally {
				loading = false;
			}
		}, 250);
		return () => clearTimeout(debounceTimer);
	});

	$effect(() => {
		if (!open) {
			query = '';
			apiResults = { artists: [], posts: [], articles: [], events: [] };
		}
	});

	function selectResult(href: string) {
		setOpen(false);
		query = '';
		goto(href);
	}

	$effect(() => {
		const onKeyDown = (e: KeyboardEvent) => {
			if (!(e.metaKey || e.ctrlKey) || e.key.toLowerCase() !== 'k') return;
			if (e.altKey || e.shiftKey) return;
			e.preventDefault();
			setOpen(!open);
		};
		window.addEventListener('keydown', onKeyDown);
		return () => window.removeEventListener('keydown', onKeyDown);
	});
</script>

{@render children()}

{#if open}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-start justify-center bg-ink/40 px-4 pt-[12vh] backdrop-blur-sm"
		onclick={() => setOpen(false)}
		role="presentation"
	>
		<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
		<div
			class="w-full max-w-lg overflow-hidden rounded-lg border border-border bg-background shadow-2xl"
			onclick={(e) => e.stopPropagation()}
			role="dialog"
			aria-modal="true"
			aria-label="Search"
			tabindex="-1"
		>
			<div class="border-b border-border px-4 py-3">
				<input
					type="search"
					bind:value={query}
					placeholder="Search artists, artwork, wiki, events… ({searchShortcutLabel()} to toggle)"
					class="w-full bg-transparent font-mono text-sm text-foreground placeholder:text-muted-foreground focus:outline-none"
				/>
			</div>
			<div class="max-h-[50vh] overflow-y-auto p-2">
				{#if useApi}
					{#if loading}
						<p class="px-3 py-6 text-center text-sm text-muted-foreground">Searching…</p>
					{:else if query.trim() && !apiHasResults}
						<p class="px-3 py-6 text-center text-sm text-muted-foreground">No results found.</p>
					{:else}
						{#if apiResults.artists.length}
							<p class="px-3 py-2 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
								Artists
							</p>
							{#each apiResults.artists as artist}
								<button
									type="button"
									class="flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left transition hover:bg-muted"
									onclick={() => selectResult(`/artists/${artist.slug}`)}
								>
									<span class="min-w-0 flex-1 font-display text-sm">{artist.display_name}</span>
									<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
										{artist.discipline?.split(' / ')[0] ?? ''}
									</span>
								</button>
							{/each}
						{/if}
						{#if apiResults.posts.length}
							<p class="px-3 py-2 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
								Artwork
							</p>
							{#each apiResults.posts as post}
								<button
									type="button"
									class="flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left transition hover:bg-muted"
									onclick={() => selectResult(`/artists/${post.artist_slug}`)}
								>
									<span class="min-w-0 flex-1 font-display text-sm">{post.title}</span>
									<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
										{post.artist_name ?? ''}
									</span>
								</button>
							{/each}
						{/if}
						{#if apiResults.articles.length}
							<p class="px-3 py-2 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
								Wiki
							</p>
							{#each apiResults.articles as article}
								<button
									type="button"
									class="flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left transition hover:bg-muted"
									onclick={() => selectResult(`/wiki/${article.slug}`)}
								>
									<span class="min-w-0 flex-1 font-display text-sm">{article.title}</span>
									<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
										{article.category ?? ''}
									</span>
								</button>
							{/each}
						{/if}
						{#if apiResults.events.length}
							<p class="px-3 py-2 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
								Events
							</p>
							{#each apiResults.events as event}
								<button
									type="button"
									class="flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left transition hover:bg-muted"
									onclick={() => selectResult(`/events/${event.slug}`)}
								>
									<span class="min-w-0 flex-1 font-display text-sm">{event.title}</span>
									<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
										{event.city ?? ''}
									</span>
								</button>
							{/each}
						{/if}
					{/if}
				{:else if staticFiltered.length === 0}
					<p class="px-3 py-6 text-center text-sm text-muted-foreground">No results found.</p>
				{:else}
					{#each [...staticGrouped.entries()] as [group, items]}
						<p class="px-3 py-2 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
							{group}
						</p>
						{#each items as item}
							<button
								type="button"
								class="flex w-full items-center gap-3 rounded-sm px-3 py-2 text-left transition hover:bg-muted"
								onclick={() => selectResult(item.href)}
							>
								<span class="min-w-0 flex-1 font-display text-sm">{item.label}</span>
								<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
									{item.meta}
								</span>
							</button>
						{/each}
					{/each}
				{/if}
			</div>
		</div>
	</div>
{/if}
