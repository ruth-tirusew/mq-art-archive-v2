<script lang="ts">
	import SectionEyebrow from '$lib/components/home/SectionEyebrow.svelte';
	import FeaturedArtistHero from '$lib/components/home/FeaturedArtistHero.svelte';
	import HomeFallbackHero from '$lib/components/home/HomeFallbackHero.svelte';
	import EditorialCanvas from '$lib/components/home/EditorialCanvas.svelte';
	import MarqueeStrip from '$lib/components/home/MarqueeStrip.svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import {
		artistDiscipline,
		artistLocation,
		artistName,
		artistPortrait,
		artistSlug
	} from '$lib/utils/fields';
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

	// A short roster preview for the discovery section — the featured artist already gets
	// their own hero above, so exclude them here rather than repeat the same profile twice.
	const rosterPreview = $derived(
		(data.artists ?? []).filter((a: ArtistProfile) => a.slug !== featuredArtist?.slug).slice(0, 6)
	);
	const artistsTotal = $derived(data.artistsTotal ?? (data.artists ?? []).length);
	const disciplines = $derived(
		Array.from(
			new Set(
				(data.artists ?? [])
					.map((a: ArtistProfile) => artistDiscipline(a))
					.filter((d: string | undefined): d is string => Boolean(d))
			)
		).slice(0, 6)
	);
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
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
		<SectionEyebrow number="02" label="Discover the roster" />
		<div class="mt-8 grid gap-10 md:grid-cols-12">
			<div class="md:col-span-5">
				<h2 class="max-w-md font-display text-3xl leading-[1.05] text-foreground md:text-4xl">
					Every artist, <em class="italic">searchable</em>.
				</h2>
				<p class="mt-4 max-w-md text-sm leading-relaxed text-muted-foreground md:text-base">
					Filter by discipline, sort by newest or name, and open a full profile with works, bio
					and contact details — no gatekeeping, no submission fee.
				</p>

				<dl class="mt-8 flex gap-8 border-t border-border/70 pt-6">
					<div>
						<dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
							Artists listed
						</dt>
						<dd class="mt-1 font-display text-3xl text-accent md:text-4xl">{artistsTotal}</dd>
					</div>
					{#if disciplines.length > 0}
						<div class="min-w-0">
							<dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
								Disciplines
							</dt>
							<dd class="mt-2 flex flex-wrap gap-1.5">
								{#each disciplines as d}
									<span
										class="rounded-full border border-border/70 px-2.5 py-0.5 font-mono text-[10px] uppercase tracking-[0.15em] text-foreground/80"
									>
										{d}
									</span>
								{/each}
							</dd>
						</div>
					{/if}
				</dl>

				<a
					href="/artists"
					class="mt-8 inline-block font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
				>
					Explore all artists →
				</a>
			</div>

			<div class="md:col-span-7">
				{#if rosterPreview.length > 0}
					<div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
						{#each rosterPreview as artist (artistSlug(artist))}
							<a
								href="/artists/{artistSlug(artist)}"
								class="group rounded-sm border border-border/70 bg-card/40 p-4 transition hover:border-accent"
							>
								<div class="aspect-square w-full overflow-hidden rounded-sm bg-muted">
									{#if artistPortrait(artist)}
										<img
											src={artistPortrait(artist)}
											alt={artistName(artist)}
											class="h-full w-full object-cover transition duration-500 group-hover:scale-105"
										/>
									{/if}
								</div>
								<p
									class="mt-3 truncate font-display text-base leading-tight text-foreground group-hover:text-accent"
								>
									{artistName(artist)}
								</p>
								{#if artistDiscipline(artist)}
									<p class="mt-0.5 truncate text-xs text-muted-foreground">
										{artistDiscipline(artist)}
									</p>
								{:else if artistLocation(artist)}
									<p class="mt-0.5 truncate text-xs text-muted-foreground">
										{artistLocation(artist)}
									</p>
								{/if}
							</a>
						{/each}
					</div>
				{:else}
					<div
						class="flex min-h-[220px] items-center justify-center rounded-sm border border-border/70 bg-card/40 p-8 text-center"
					>
						<p class="max-w-xs text-sm leading-relaxed text-muted-foreground">
							The artist roster is temporarily unavailable — the full directory is still at
							<a href="/artists" class="text-accent underline underline-offset-4">/artists</a>.
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
						class="mx-auto flex min-h-[280px] max-w-md items-center justify-center rounded-sm border border-border/70 bg-card/40 p-8 text-center"
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
