<script lang="ts">
	import SectionEyebrow from '$lib/components/home/SectionEyebrow.svelte';
	import FeaturedArtistHero from '$lib/components/home/FeaturedArtistHero.svelte';
	import HomeFallbackHero from '$lib/components/home/HomeFallbackHero.svelte';
	import EditorialCanvas from '$lib/components/home/EditorialCanvas.svelte';
	import MarqueeStrip from '$lib/components/home/MarqueeStrip.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const editorialSpreads = $derived(data.editorialSpreads ?? data.editorialWalls ?? []);
	const featuredArtist = $derived(data.featuredArtist);
	const wikiPreview = $derived((data.articles ?? []).slice(0, 3));
	const marqueeItems = $derived(data.marqueeItems ?? []);

	function articleExcerpt(article: Article): string {
		return article.excerpt ?? article.body.slice(0, 120);
	}

	function articleCategory(article: Article): string {
		return article.category ?? 'General';
	}
</script>

<svelte:head>
	<title>Artiv — Discover Ethiopian artists</title>
	<meta
		name="description"
		content="Discover Ethiopian artists and explore their worlds — a living archive of modern Ethiopian art and its diaspora."
	/>
</svelte:head>

{#if data.showEditorialHero}
	<EditorialCanvas spreads={editorialSpreads} />
{:else}
	<HomeFallbackHero />
{/if}

{#if featuredArtist}
	<FeaturedArtistHero artist={featuredArtist} posts={data.featuredPosts ?? []} />
{/if}

<MarqueeStrip items={marqueeItems} />

<section class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
	<SectionEyebrow number="02" label="Held in common" />
	<div class="mt-8 grid gap-4 md:grid-cols-12">
		{#if wikiPreview.length > 0}
			<div class="grid gap-4 md:col-span-7 md:grid-cols-1">
				{#each wikiPreview as a}
					<a
						href="/wiki/{a.slug}"
						class="group rounded-sm border border-border/70 bg-card/40 p-6 transition hover:border-accent"
					>
						<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
							{articleCategory(a)}
						</p>
						<h3
							class="mt-2 font-display text-xl leading-tight text-foreground group-hover:text-accent md:text-2xl"
						>
							{a.title}
						</h3>
						<p class="mt-2 text-sm leading-relaxed text-muted-foreground">{articleExcerpt(a)}</p>
					</a>
				{/each}
				<a
					href="/wiki"
					class="font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
				>
					Browse all wiki entries →
				</a>
			</div>
		{:else}
			<a
				href="/wiki"
				class="group relative flex flex-col justify-between overflow-hidden rounded-sm border border-border/70 bg-card/40 p-8 transition hover:border-accent md:col-span-7 md:min-h-[280px]"
			>
				<div>
					<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
						⁂ Wiki · Community handbook
					</p>
					<h3 class="mt-4 max-w-lg font-display text-3xl leading-tight text-foreground md:text-4xl">
						The handbook for working as an artist in Ethiopia.
					</h3>
					<p class="mt-4 max-w-md text-sm leading-relaxed text-muted-foreground">
						Crowdsourced entries on EIPA registration, gallery contracts, pigment sources and payment
						rails — written by people who file the paperwork themselves.
					</p>
				</div>
				<span
					class="mt-6 font-mono text-[11px] uppercase tracking-[0.2em] text-foreground transition group-hover:tracking-[0.3em] group-hover:text-accent"
				>
					Read the wiki →
				</span>
			</a>
		{/if}

		<a
			href="/events"
			class="group relative flex flex-col justify-between overflow-hidden rounded-sm border border-border/70 bg-ink p-8 text-cream transition hover:border-accent md:col-span-5 md:min-h-[280px]"
		>
			<div>
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
					✺ Events · This week
				</p>
				<h3 class="mt-4 font-display text-3xl leading-tight md:text-4xl">What's opening in Addis.</h3>
				<p class="mt-4 text-sm leading-relaxed text-cream/75">
					A curated calendar of openings, talks and residencies — verified by institutional partners.
				</p>
			</div>
			<span
				class="mt-6 font-mono text-[11px] uppercase tracking-[0.2em] text-cream transition group-hover:tracking-[0.3em] group-hover:text-accent"
			>
				See calendar →
			</span>
		</a>
	</div>
</section>

<section class="border-t border-border/60 bg-card/30">
	<div class="mx-auto flex max-w-[1600px] flex-wrap items-center justify-between gap-6 px-6 py-8 md:px-10">
		<p class="max-w-xl font-display text-lg text-foreground md:text-xl">
			Are you an artist?
			<span class="text-muted-foreground">Claim a shareable profile.</span>
		</p>
		<a
			href="/portfolio"
			class="font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
		>
			Portfolio builder →
		</a>
	</div>
</section>
