<script lang="ts">
	import ArtworkCard from '$lib/components/art/ArtworkCard.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import { acquisitionArtistSlug } from '$lib/utils/fields';
	import { Button, EmptyState, PageHeader } from '$ui';
	import type { PageData } from './$types';

	type FilterKey = 'city' | 'medium' | 'year' | 'style' | 'residency' | 'exhibition';

	let { data }: { data: PageData } = $props();

	const items = $derived(data.posts as ArtPost[]);

	const filterGroups: { key: FilterKey; label: string; options: (string | number)[] }[] = $derived([
		{ key: 'city', label: 'City', options: data.cities },
		{ key: 'medium', label: 'Medium', options: data.mediums },
		{ key: 'year', label: 'Year', options: data.years },
		{ key: 'style', label: 'Style', options: data.styles },
		{ key: 'residency', label: 'Residency', options: data.residencies },
		{ key: 'exhibition', label: 'Exhibition', options: data.exhibitions }
	]);

	let active = $state<Partial<Record<FilterKey, string>>>({});

	const filtered = $derived(
		items.filter((w) => {
			if (active.city && w.city !== active.city) return false;
			if (active.medium && !w.medium?.startsWith(active.medium)) return false;
			if (active.year && w.year !== Number(active.year)) return false;
			if (active.style && w.style !== active.style) return false;
			if (active.residency && w.residency !== active.residency) return false;
			if (active.exhibition && w.exhibition !== active.exhibition) return false;
			return true;
		})
	);

	const hasFilters = $derived(Object.keys(active).length > 0);

	function toggle(key: FilterKey, value: string) {
		const next = { ...active };
		if (next[key] === value) delete next[key];
		else next[key] = value;
		active = next;
	}
</script>

<svelte:head>
	<title>Archive — Artiv</title>
	<meta
		name="description"
		content="Every work currently held in the Artiv archive of modern Ethiopian art."
	/>
</svelte:head>

<section class="container-page py-14 md:py-20">
	<PageHeader
		eyebrow="All works · {items.length}"
		lead="Browse the full collection. Filter by city, medium, year, or style."
	>
		The <em class="italic">archive</em>
	</PageHeader>

	{#if items.length === 0}
		<EmptyState class="mt-10" title="No works have been added yet." body="Works appear here once artists publish them from their studio.">
			{#snippet action()}
				<Button variant="secondary" size="sm" href="/artists">Meet the artists →</Button>
				<Button variant="link" href="/apply">Apply to be listed →</Button>
			{/snippet}
		</EmptyState>
	{:else}
	<div
		class="sticky top-[65px] z-30 -mx-6 mt-10 border-y border-border/60 bg-background/95 px-6 py-4 backdrop-blur md:-mx-10 md:px-10"
	>
		<div class="flex flex-wrap items-center gap-x-6 gap-y-4">
			{#each filterGroups as group}
				<div class="flex flex-wrap items-center gap-2">
					<span class="font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
						{group.label}
					</span>
					{#each group.options as opt}
						{@const val = String(opt)}
						<button
							type="button"
							onclick={() => toggle(group.key, val)}
							class="rounded-full border px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.15em] transition {active[
								group.key
							] === val
								? 'border-foreground bg-foreground text-background'
								: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
						>
							{val}
						</button>
					{/each}
				</div>
			{/each}
			{#if hasFilters}
				<button
					type="button"
					onclick={() => (active = {})}
					class="ml-auto font-mono text-[9px] uppercase tracking-[0.2em] text-accent hover:underline"
				>
					Clear filters
				</button>
			{/if}
		</div>
	</div>

	<p class="t-meta mt-6 text-muted-foreground">
		Showing {filtered.length} of {items.length}
	</p>

	<div class="mt-8 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-3">
		{#each filtered as w (w.id)}
			<ArtworkCard item={w} href="/artists/{acquisitionArtistSlug(w)}" showArtist />
		{/each}
	</div>

	{#if filtered.length === 0}
		<EmptyState class="mt-10" eyebrow="No matches" title="No works match these filters." body="Try removing a filter to see more of the archive.">
			{#snippet action()}<Button variant="secondary" size="sm" onclick={() => (active = {})}>Clear filters</Button>{/snippet}
		</EmptyState>
	{/if}
	{/if}
</section>
