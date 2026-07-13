<script lang="ts">
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

	type SortKey = 'name' | 'works' | 'recent';

	let { data }: { data: PageData } = $props();

	const roster = $derived(data.artists);
	const postsBySlug = $derived(data.postsBySlug);
	const totalWorks = $derived(data.totalWorks);

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

	let filter = $state<string | null>(null);
	let sort = $state<SortKey>('name');

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
	<title>Artists — Mäkdäs</title>
	<meta
		name="description"
		content="The painters, printmakers, and image-keepers in the Mäkdäs archive."
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
<section class="border-b border-border/60 bg-ink text-cream">
	<div class="mx-auto grid max-w-[1600px] grid-cols-12 gap-6 px-6 py-14 md:px-10 md:py-20">
		<div class="col-span-12 mb-2 md:col-span-7">
			<p class="flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.3em] text-cream/60">
				<span>✕</span> Index — {roster.length} artists · {totalWorks} works
			</p>
		</div>
		<a href="/artists/{featured.slug}" class="group col-span-12 md:col-span-5">
			<div class="grain relative aspect-[4/5] overflow-hidden rounded-sm bg-card">
				{#if featuredCover}
					<img
						src={featuredCover}
						alt={artistName(featured)}
						class="h-full w-full object-cover transition duration-700 group-hover:scale-[1.03]"
					/>
				{/if}
				<span
					class="absolute left-4 top-4 rounded-full bg-accent px-3 py-1 font-mono text-[9px] uppercase tracking-[0.25em] text-accent-foreground"
				>
					● Featured
				</span>
			</div>
		</a>
		<div class="col-span-12 flex flex-col justify-between md:col-span-7">
			<div>
				<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-cream/60">
					Spotlight / Q{quarter}
				</p>
				<a href="/artists/{featured.slug}" class="group block">
					<h2
						class="mt-4 font-display text-5xl leading-[0.95] transition group-hover:text-accent md:text-7xl"
					>
						{artistName(featured)}
					</h2>
				</a>
				<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/50">
					@{artistHandle(featured)}
				</p>
				{#if artistTagline(featured)}
					<p class="mt-6 max-w-lg text-balance font-display text-xl italic text-cream/85 md:text-2xl">
						"{artistTagline(featured)}"
					</p>
				{/if}
			</div>
			<div class="mt-10 flex flex-wrap items-end justify-between gap-6">
				<div
					class="grid grid-cols-2 gap-x-8 gap-y-3 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70"
				>
					<div><span class="text-cream/40">Based · </span>{artistLocation(featured)}</div>
					<div><span class="text-cream/40">Active · </span>{artistYearsActive(featured) ?? '—'}</div>
					<div><span class="text-cream/40">Works · </span>{worksFor(featured.slug).length}</div>
					<div><span class="text-cream/40">Discipline · </span>{artistDiscipline(featured) ?? '—'}</div>
				</div>
				<div class="flex flex-wrap items-center gap-4">
					<a
						href="/@{artistHandle(featured)}"
						class="font-mono text-[11px] uppercase tracking-[0.25em] text-cream/70 underline decoration-cream/30 underline-offset-4 transition hover:text-accent hover:decoration-accent"
					>
						Share →
					</a>
					<a
						href="/artists/{featured.slug}"
						class="font-mono text-[11px] uppercase tracking-[0.25em] text-accent transition hover:tracking-[0.35em]"
					>
						Open profile →
					</a>
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
				onclick={() => (filter = null)}
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
						onclick={() => (filter = filter === d ? null : d)}
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
					onclick={() => (sort = k as SortKey)}
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

<section class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
	<div class="grid auto-rows-[minmax(240px,auto)] grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-6">
		{#each filtered as a, i}
			{@const w = worksFor(a.slug)}
			{@const span = layouts[i % layouts.length]}
			{@const isLarge = span.includes('row-span-2')}
			{@const cover = coverFor(a.slug, a)}
			{@const handle = artistHandle(a)}
			<div
				class="group relative flex flex-col overflow-hidden rounded-sm border border-border/60 bg-card/40 transition hover:border-accent {span}"
			>
				<a href="/artists/{a.slug}" class="flex min-h-0 flex-1 flex-col">
					<div class="relative overflow-hidden {isLarge ? 'flex-1' : 'aspect-[4/3]'} bg-muted">
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
							class="absolute right-3 top-3 rounded-full bg-background/85 px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.2em] text-foreground backdrop-blur"
						>
							{w.length} works
						</span>
						{#if isLarge && w[0]}
							<div class="absolute bottom-3 left-3 flex gap-1">
								{#each acquisitionPalette(w[0]).slice(0, 4) as c}
									<span class="h-5 w-1.5" style:background={c}></span>
								{/each}
							</div>
						{/if}
					</div>
					<div class="flex items-start justify-between gap-3 border-t border-border/60 p-4">
						<div class="min-w-0">
							<p
								class="font-display text-foreground {isLarge
									? 'text-2xl'
									: 'text-lg'} leading-tight"
							>
								{artistName(a)}
							</p>
							<p class="mt-1 truncate font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
								{artistDiscipline(a)} · {artistLocation(a)}
							</p>
							<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground/70">
								@{handle}
							</p>
						</div>
						<span
							class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground transition group-hover:translate-x-1 group-hover:text-accent"
						>
							→
						</span>
					</div>
					{#if isLarge && artistTagline(a)}
						<p class="border-t border-border/60 px-4 py-3 font-display text-sm italic text-foreground/75">
							"{artistTagline(a)}"
						</p>
					{/if}
				</a>
				<div class="border-t border-border/60 px-4 py-2">
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
				href="/about"
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
</section>
