<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { ArtPost } from '$lib/core/domain/art';
	import {
		acquisitionImage,
		acquisitionPalette,
		artistDiscipline,
		artistHandle,
		artistLocation,
		artistName,
		artistPortrait,
		artistTagline,
		artistYearsActive
	} from '$lib/utils/fields';
	import type { PageData } from './$types';
	import type { ArtistsSortKey } from './+page';

	let { data }: { data: PageData } = $props();

	const roster = $derived(data.artists);
	const postsBySlug = $derived(data.postsBySlug);
	const totalWorks = $derived(data.totalWorks);
	const loadFailed = $derived(data.loadFailed);
	const pagination = $derived(data.pagination);

	function worksFor(slug: string): ArtPost[] {
		return postsBySlug[slug] ?? [];
	}

	function latestWorkYear(slug: string): number {
		return worksFor(slug).reduce((max, work) => Math.max(max, work.year ?? 0), 0);
	}

	function coverFor(slug: string, artist: (typeof roster)[number]): string | undefined {
		const works = worksFor(slug);
		const fromWork = works[0] ? acquisitionImage(works[0]) : undefined;
		return fromWork ?? artistPortrait(artist);
	}

	const disciplines = $derived(
		Array.from(new Set(roster.map((a) => artistDiscipline(a)).filter((d): d is string => Boolean(d))))
	);

	const filter = $derived(data.filters.discipline);
	const sort = $derived(data.filters.sort);

	function syncQuery(next: { discipline?: string | null; sort?: ArtistsSortKey }) {
		const params = new URLSearchParams($page.url.searchParams);
		const discipline = next.discipline !== undefined ? next.discipline : filter;
		const nextSort = next.sort !== undefined ? next.sort : sort;

		if (discipline) params.set('discipline', discipline);
		else params.delete('discipline');

		if (nextSort && nextSort !== 'name') params.set('sort', nextSort);
		else params.delete('sort');
		params.delete('page');

		const qs = params.toString();
		goto(qs ? `/artists?${qs}` : '/artists', {
			replaceState: true,
			keepFocus: true,
			noScroll: true
		});
	}

	function setFilter(value: string | null) {
		syncQuery({ discipline: value });
	}

	function setSort(value: ArtistsSortKey) {
		syncQuery({ sort: value });
	}

	const featured = $derived(roster.find((a) => a.featured) ?? null);
	const featuredCover = $derived(featured ? coverFor(featured.slug, featured) : undefined);
	const rest = $derived(roster.filter((a) => a.slug !== featured?.slug));

	const filtered = $derived.by(() => {
		const list = filter ? rest.filter((a) => artistDiscipline(a) === filter) : rest;
		return [...list].sort((a, b) => {
			if (sort === 'name') return artistName(a).localeCompare(artistName(b));
			if (sort === 'works') return worksFor(b.slug).length - worksFor(a.slug).length;
			return latestWorkYear(b.slug) - latestWorkYear(a.slug);
		});
	});

	const layouts = [
		'lg:col-span-3 lg:row-span-2',
		'lg:col-span-3',
		'lg:col-span-2',
		'lg:col-span-2',
		'lg:col-span-2'
	];

	const updated = new Date().toLocaleDateString('en-GB', { month: 'short', year: 'numeric' });
	const quarter = Math.ceil((new Date().getMonth() + 1) / 3);
</script>

<svelte:head>
	<title>Artists — Artiv</title>
	<meta
		name="description"
		content="The painters, printmakers, and image-keepers in the Artiv archive."
	/>
</svelte:head>

{#if !featured}
<section class="border-b border-border/60">
	<div class="mx-auto grid max-w-[1600px] grid-cols-12 gap-6 px-6 py-14 md:px-10 md:py-20">
		<div class="col-span-12 md:col-span-8">
			<p class="flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
				<span>✕</span> Index — {roster.length} artists · {totalWorks} works
			</p>
			<h1 class="mt-6 font-display text-[13vw] leading-[0.9] tracking-tight text-foreground md:text-[7.5vw]">
				The <em class="italic text-accent">roster</em>.
			</h1>
			<p class="mt-6 max-w-xl text-balance text-lg leading-relaxed text-muted-foreground">
				Painters, printmakers, and image-keepers working out of Addis, Bahir Dar, Berlin and Brooklyn
				— held together by a shared inheritance and a stubborn refusal to inherit quietly.
			</p>
		</div>
		<aside
			class="col-span-12 space-y-2 border-l border-border/60 pl-6 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground md:col-span-4"
		>
			<p class="text-foreground">Legend</p>
			<p><span class="text-accent">●</span> Featured this quarter</p>
			<p><span class="text-secondary">◐</span> In residence abroad</p>
			<p>▢ &nbsp; Open for commission</p>
			<p class="pt-3 text-muted-foreground/70">Updated {updated}</p>
		</aside>
	</div>
</section>
{/if}

{#if featured}
<section class="max-h-[calc(100vh-65px)] overflow-hidden border-b border-border/60 bg-ink text-cream">
	<div
		class="mx-auto grid max-h-[calc(100vh-65px)] max-w-[1600px] grid-cols-12 gap-4 px-6 py-8 md:grid-rows-[auto_minmax(0,1fr)] md:gap-5 md:px-10 md:py-8"
	>
		<div class="col-span-12 md:col-span-7 md:row-start-1">
			<p class="flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.3em] text-cream/60">
				<span>✕</span> Index — {roster.length} artists · {totalWorks} works
			</p>
		</div>
		<a
			href="/artists/{featured.slug}"
			class="group col-span-12 min-h-0 md:col-span-5 md:row-span-2 md:row-start-1"
		>
			<div
				class="grain relative aspect-[4/5] max-h-[min(56vh,520px)] overflow-hidden rounded-sm bg-card md:aspect-auto md:h-full md:max-h-none"
			>
				{#if featuredCover}
					<img
						src={featuredCover}
						alt={artistName(featured)}
						class="h-full w-full object-cover transition duration-700 group-hover:scale-[1.03]"
					/>
				{/if}
				<span
					class="absolute left-3 top-3 rounded-full bg-accent px-3 py-1 font-mono text-[9px] uppercase tracking-[0.25em] text-accent-foreground"
				>
					● Featured
				</span>
			</div>
		</a>
		<div class="col-span-12 flex min-h-0 flex-col justify-between md:col-span-7 md:row-start-2">
			<div>
				<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-cream/60">
					Spotlight / Q{quarter}
				</p>
				<a href="/artists/{featured.slug}" class="group block">
					<h2
						class="mt-2 font-display text-4xl leading-[0.95] transition group-hover:text-accent md:text-5xl lg:text-6xl"
					>
						{artistName(featured)}
					</h2>
				</a>
				<p class="mt-1.5 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/50">
					@{artistHandle(featured)}
				</p>
				{#if artistTagline(featured)}
					<p class="mt-3 max-w-lg text-balance font-display text-lg italic text-cream/85 md:text-xl">
						"{artistTagline(featured)}"
					</p>
				{/if}
			</div>
			<div class="mt-5 flex flex-wrap items-end justify-between gap-4">
				<div
					class="grid grid-cols-2 gap-x-8 gap-y-2 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70"
				>
					<div><span class="text-cream/40">Based · </span>{artistLocation(featured)}</div>
					<div><span class="text-cream/40">Active · </span>{artistYearsActive(featured) ?? '—'}</div>
					<div><span class="text-cream/40">Works · </span>{worksFor(featured.slug).length}</div>
					<div><span class="text-cream/40">Discipline · </span>{artistDiscipline(featured) ?? '—'}</div>
				</div>
			</div>
		</div>
	</div>
</section>
{/if}

<section class="sticky top-[64px] z-10 border-b border-border/60 bg-background/90 backdrop-blur">
	<div class="mx-auto flex max-w-[1600px] flex-wrap items-center justify-between gap-4 px-6 py-3 md:px-10">
		<div class="flex flex-wrap items-center gap-2 font-mono text-[10px] uppercase tracking-[0.2em]">
			<button
				type="button"
				onclick={() => setFilter(null)}
				class="rounded-full border px-3 py-1 transition {filter === null
					? 'border-foreground bg-foreground text-background'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				All ({rest.length})
			</button>
			{#each disciplines as d}
				{@const count = rest.filter((a) => artistDiscipline(a) === d).length}
				{#if count}
					<button
						type="button"
						onclick={() => setFilter(filter === d ? null : d)}
						class="rounded-full border px-3 py-1 transition {filter === d
							? 'border-accent bg-accent text-accent-foreground'
							: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
					>
						{d} ({count})
					</button>
				{/if}
			{/each}
		</div>
		<div class="flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
			<span>Sort</span>
			{#each ['name', 'works', 'recent'] as k}
				<button
					type="button"
					onclick={() => setSort(k as ArtistsSortKey)}
					class="transition {sort === k
						? 'text-foreground underline decoration-accent underline-offset-4'
						: 'hover:text-foreground'}"
				>
					{k === 'name' ? 'A → Z' : k === 'works' ? 'Most works' : 'Most recent'}
				</button>
			{/each}
		</div>
	</div>
</section>

<section class="mx-auto max-w-[1600px] px-6 py-10 md:px-10 md:py-14">
	{#if loadFailed && roster.length === 0}
		<p class="py-16 text-center font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
			Roster unavailable right now. Try again shortly.
		</p>
	{:else}
		<div class="grid auto-rows-[minmax(160px,auto)] grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-6">
			{#each filtered as a, i}
				{@const w = worksFor(a.slug)}
				{@const span = layouts[i % layouts.length]}
				{@const isLarge = span.includes('row-span-2')}
				{@const cover = coverFor(a.slug, a)}
				{@const handle = artistHandle(a)}
				<div
					data-testid="web-artist-card"
					class="group relative flex flex-col overflow-hidden rounded-sm border border-border/60 bg-card/40 transition hover:border-accent {span}"
				>
					<a href="/artists/{a.slug}" class="flex min-h-0 flex-1 flex-col">
						<div class="relative overflow-hidden {isLarge ? 'flex-1' : 'aspect-[5/3]'} bg-muted">
							{#if cover}
								<img
									src={cover}
									alt={artistName(a)}
									loading="lazy"
									class="h-full w-full object-cover grayscale transition duration-700 group-hover:scale-[1.04] group-hover:grayscale-0"
								/>
							{/if}
							<div
								class="absolute inset-0 bg-gradient-to-t from-ink/60 via-transparent to-transparent opacity-0 transition group-hover:opacity-100"
							></div>
							<span
								class="absolute right-2.5 top-2.5 rounded-full bg-background/85 px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.2em] text-foreground backdrop-blur"
							>
								{w.length} works
							</span>
							{#if isLarge && w[0]}
								<div class="absolute bottom-2.5 left-2.5 flex gap-1">
									{#each acquisitionPalette(w[0]).slice(0, 4) as c}
										<span class="h-4 w-1.5" style:background={c}></span>
									{/each}
								</div>
							{/if}
						</div>
						<div class="flex items-start justify-between gap-3 border-t border-border/60 px-3 py-2.5">
							<div class="min-w-0">
								<p
									class="font-display text-foreground {isLarge
										? 'text-xl'
										: 'text-base'} leading-tight"
								>
									{artistName(a)}
								</p>
							<p class="mt-0.5 truncate font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
								{artistDiscipline(a)} · {artistLocation(a)}
							</p>
							<p class="mt-0.5 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground/70">
								@{handle}
								{#if a.in_residence}
									<span class="text-secondary"> · ◐</span>
								{/if}
								{#if a.open_for_commission}
									<span> · ▢</span>
								{/if}
							</p>
							</div>
							<span
								class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground transition group-hover:translate-x-1 group-hover:text-accent"
							>
								→
							</span>
						</div>
						{#if isLarge && artistTagline(a)}
							<p class="border-t border-border/60 px-3 py-2 font-display text-sm italic text-foreground/75">
								"{artistTagline(a)}"
							</p>
						{/if}
					</a>
					<div class="border-t border-border/60 px-3 py-1.5">
						<a
							href="/@{handle}"
							class="font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground transition hover:text-accent"
						>
							Share profile →
						</a>
					</div>
				</div>
			{/each}

			<div
				class="flex flex-col justify-between rounded-sm border border-dashed border-border p-5 sm:col-span-2 lg:col-span-2"
			>
				<div>
					<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">⁂ Open call</p>
					<p class="mt-3 font-display text-xl leading-tight text-foreground">
						Nominate an artist for the archive.
					</p>
					<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
						Tier-2 members can propose a file. Tier-3 institutional partners publish directly.
					</p>
				</div>
				<a
					href="/apply"
					class="mt-6 inline-block font-mono text-[10px] uppercase tracking-[0.25em] text-foreground underline decoration-accent underline-offset-8"
				>
					How vetting works →
				</a>
			</div>
		</div>

		{#if filtered.length === 0}
			<p class="mt-16 text-center font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
				No artists match this filter yet.
			</p>
		{/if}
		{#if pagination.pages > 1}
			<nav class="mt-12 flex items-center justify-center gap-4 font-mono text-[10px] uppercase tracking-[0.2em]" aria-label="Artist pages">
				{#if pagination.page > 1}
					<a href="?{new URLSearchParams({ ...Object.fromEntries($page.url.searchParams), page: String(pagination.page - 1) })}">← Previous</a>
				{:else}<span class="opacity-40">← Previous</span>{/if}
				<span>Page {pagination.page} of {pagination.pages}</span>
				{#if pagination.page < pagination.pages}
					<a href="?{new URLSearchParams({ ...Object.fromEntries($page.url.searchParams), page: String(pagination.page + 1) })}">Next →</a>
				{:else}<span class="opacity-40">Next →</span>{/if}
			</nav>
		{/if}
	{/if}
</section>
