<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import type { Work } from '$lib/data/archive';
	import {
		acquisitionArtistName,
		acquisitionArtistSlug,
		acquisitionImage,
		acquisitionPalette,
		acquisitionTitle
	} from '$lib/utils/fields';
	import type { PageData } from './$types';

	type FilterKey = 'city' | 'medium' | 'year' | 'style' | 'residency' | 'exhibition';

	let { data }: { data: PageData } = $props();

	const items = $derived(data.posts as (Work | ArtPost)[]);

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
			if (active.residency && 'residency' in w && w.residency !== active.residency) return false;
			if (active.exhibition && 'exhibition' in w && w.exhibition !== active.exhibition) return false;
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
	<title>Archive — Mäkdäs</title>
	<meta
		name="description"
		content="Every work currently held in the Mäkdäs archive of modern Ethiopian art."
	/>
</svelte:head>

<section class="mx-auto max-w-[1600px] px-6 py-14 md:px-10 md:py-20">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-terracotta">
		⁂ &nbsp; All works · {items.length}
	</p>
	<h1 class="mt-4 font-display text-4xl text-foreground md:text-6xl">The archive</h1>
	<p class="mt-4 max-w-2xl text-muted-foreground">
		Browse the full collection. Filter by city, medium, year, style, residency, or exhibition.
	</p>

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

	<p class="mt-6 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
		Showing {filtered.length} of {items.length}
	</p>

	<div class="mt-8 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-3">
		{#each filtered as w}
			{@const image = acquisitionImage(w)}
			<a href="/artists/{acquisitionArtistSlug(w)}" class="group">
				<div class="grain relative aspect-[4/5] overflow-hidden rounded-sm bg-card">
					{#if image}
						<img
							src={image}
							alt={acquisitionTitle(w)}
							loading="lazy"
							class="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.06]"
						/>
					{/if}
					<div class="absolute inset-0 bg-ink/0 transition duration-500 group-hover:bg-ink/25"></div>
					<div
						class="absolute inset-x-0 bottom-0 translate-y-3 p-4 opacity-0 transition duration-500 group-hover:translate-y-0 group-hover:opacity-100"
					>
						<div class="bg-gradient-to-t from-ink/95 via-ink/75 to-transparent p-4">
							<p class="font-display text-lg italic text-cream">{acquisitionTitle(w)}</p>
							<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/80">
								{acquisitionArtistName(w)}
							</p>
							<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/60">
								{w.medium} · {w.year}
							</p>
							<span class="mt-2 inline-block font-mono text-[10px] uppercase tracking-[0.2em] text-accent">
								View →
							</span>
						</div>
					</div>
				</div>
				<div class="mt-4 flex items-start justify-between gap-4 transition group-hover:opacity-0">
					<div>
						<p class="font-display text-lg italic text-foreground">{acquisitionTitle(w)}</p>
						<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
							{acquisitionArtistName(w)} · {w.year}
						</p>
					</div>
					<div class="flex gap-1 pt-1">
						{#each acquisitionPalette(w) as c}
							<span class="h-5 w-1.5" style:background={c}></span>
						{/each}
					</div>
				</div>
			</a>
		{/each}
	</div>

	{#if filtered.length === 0}
		<p class="mt-16 text-center text-muted-foreground">No works match these filters.</p>
	{/if}
</section>
