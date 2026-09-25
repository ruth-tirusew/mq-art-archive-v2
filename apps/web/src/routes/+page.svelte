<script lang="ts">
	import SectionEyebrow from '$lib/components/home/SectionEyebrow.svelte';
	import FeaturedArtistHero from '$lib/components/home/FeaturedArtistHero.svelte';
	import HomeFallbackHero from '$lib/components/home/HomeFallbackHero.svelte';
	import EditorialCanvas from '$lib/components/home/EditorialCanvas.svelte';
	import MarqueeStrip from '$lib/components/home/MarqueeStrip.svelte';
	import RosterRadar from '$lib/components/home/RosterRadar.svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import CtaLink from '$lib/components/CtaLink.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const heroSpreads = $derived(data.heroSpreads ?? []);
	const featuredArtist = $derived(data.featuredArtist);
	const wikiPreview = $derived((data.articles ?? []).slice(0, 3));
	const marqueeItems = $derived(data.marqueeItems ?? []);

	function articleExcerpt(article: Article): string {
		return article.excerpt ?? article.body.slice(0, 120);
	}

	function articleCategory(article: Article): string {
		return article.category ?? 'General';
	}

	// The radar's node set — the featured artist already gets their own hero above, so
	// exclude them here rather than repeat the same profile twice. Capped well below the
	// full roster so the chart reads as a sample, not a cramped attempt at completeness.
	const nodeArtists = $derived(
		(data.artists ?? []).filter((a: ArtistProfile) => a.slug !== featuredArtist?.slug).slice(0, 14)
	);
	const artistsTotal = $derived(data.artistsTotal ?? (data.artists ?? []).length);
</script>

<svelte:head>
	<title>Artiv — Discover Ethiopian artists</title>
	<meta
		name="description"
		content="Discover Ethiopian artists and explore their worlds — a living archive of modern Ethiopian art and its diaspora."
	/>
</svelte:head>

{#if heroSpreads.length > 0}
	<EditorialCanvas spreads={heroSpreads} />
{:else}
	<HomeFallbackHero />
{/if}

{#if featuredArtist}
	<FeaturedArtistHero artist={featuredArtist} posts={data.featuredPosts ?? []} />
{/if}

<MarqueeStrip items={marqueeItems} />
<section class="border-b border-border/60">
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20 lg:py-24">

		<div class="mt-10 grid items-center gap-12 lg:grid-cols-12 lg:gap-8">
			<!-- Copy -->
			<div class="lg:col-span-5 lg:pr-12">
					<SectionEyebrow number="02" label="Discover the roster" />

				<h2
					class="max-w-lg font-display text-5xl leading-[0.95] tracking-[-0.03em] text-foreground md:text-6xl"
				>
					Every artist,
					<em class="italic">searchable.</em>
				</h2>

				<p
					class="mt-6 max-w-md text-sm leading-relaxed text-muted-foreground md:text-base"
				>
					Explore the artists in our roster, discover their disciplines,
					and open a full profile with works, biography and contact details.
				</p>

				<div class="mt-8 flex flex-wrap items-center gap-6">
					<CtaLink href="/artists" variant="accent">
						Explore all artists
						<span aria-hidden="true">→</span>
					</CtaLink>

					<p
						class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
					>
						{artistsTotal} artists listed
					</p>
				</div>
			</div>

			<!-- Radar -->
			<div class="lg:col-span-7">
				{#if nodeArtists.length > 0}
					<RosterRadar artists={nodeArtists} />
				{:else}
					<div
						class="mx-auto flex min-h-55 max-w-xl items-center justify-center rounded-sm border border-border/70 bg-card/40 p-8 text-center"
					>
						<p class="max-w-xs text-sm leading-relaxed text-muted-foreground">
							The artist roster is temporarily unavailable — the full directory is still at
							<a
								href="/artists"
								class="text-accent underline underline-offset-4"
							>
								/artists
							</a>.
						</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
</section>
<section class="border-b border-border/60 bg-card/30">
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
		<SectionEyebrow number="03" label="One link, the whole studio" />
		<div class="mt-8 grid items-center gap-10 md:grid-cols-12">
			<div class="md:col-span-7">
				<h2 class="max-w-lg font-display text-3xl leading-[1.05] text-foreground md:text-4xl">
					Portfolio · <em class="italic">link in bio</em>, built for artists.
				</h2>
				<p class="mt-4 max-w-md text-sm leading-relaxed text-muted-foreground md:text-base">
					A focused, shareable profile — works, bio, Telegram, Instagram and direct contact —
					surfaced behind one @handle so a gallerist or collector can reach you in two taps.
				</p>

				<ul class="mt-8 grid grid-cols-2 gap-x-6 gap-y-3 text-sm text-foreground/90">
					{#each ['Verified handle + tier badge', 'Six-work portfolio strip', 'Telegram, IG & WhatsApp', 'Direct contact links', 'Press kit + high-res download', 'QR for in-person handoff'] as f}
						<li class="flex items-start gap-2.5">
							<span class="mt-1.5 inline-block h-1.5 w-1.5 shrink-0 rounded-full bg-accent"></span>
							{f}
						</li>
					{/each}
				</ul>

				<a
					href="/portfolio"
					class="mt-8 inline-block font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
				>
					Claim your @handle →
				</a>
			</div>

			<div class="md:col-span-5">
				{#if featuredArtist}
					<ShareableProfile artist={featuredArtist} works={data.featuredPosts ?? []} demo framed />
				{:else}
					<div
						class="mx-auto flex min-h-70 max-w-md items-center justify-center rounded-sm border border-border/70 bg-card/40 p-8 text-center"
					>
						<p class="max-w-xs text-sm leading-relaxed text-muted-foreground">
							The profile preview is temporarily unavailable — see a live example at
							<a href="/portfolio" class="text-accent underline underline-offset-4">/portfolio</a>.
						</p>
					</div>
				{/if}
			</div>
		</div>
	</div>
</section>

<section class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
	<SectionEyebrow number="04" label="Held in common" />
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
				class="group relative flex flex-col justify-between overflow-hidden rounded-sm border border-border/70 bg-card/40 p-8 transition hover:border-accent md:col-span-7 md:min-h-70"
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
			class="group relative flex flex-col justify-between overflow-hidden rounded-sm border border-border/70 bg-ink p-8 text-cream transition hover:border-accent md:col-span-5 md:min-h-70"
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
			Artist, institution or gallery in Ethiopia?
			<span class="text-muted-foreground">Get listed on Artiv.</span>
		</p>
		<a
			href="/apply"
			class="font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
		>
			Apply →
		</a>
	</div>
</section>
