<script lang="ts">
	import SectionEyebrow from '$lib/components/home/SectionEyebrow.svelte';
	import FeaturedArtistHero from '$lib/components/home/FeaturedArtistHero.svelte';
	import HomeFallbackHero from '$lib/components/home/HomeFallbackHero.svelte';
	import EmptySectionPrompt from '$lib/components/home/EmptySectionPrompt.svelte';
	import MarqueeStrip from '$lib/components/home/MarqueeStrip.svelte';
	import RecentAcquisitionsGrid from '$lib/components/home/RecentAcquisitionsGrid.svelte';
	import CtaLink from '$lib/components/CtaLink.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import type { WikiArticle } from '$lib/data/wiki';
	import {
		artistDiscipline,
		artistLocation,
		artistName,
		artistPortrait
	} from '$lib/utils/fields';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const featuredArtist = $derived(data.featuredArtist);
	const acquisitions = $derived(data.acquisitions);
	const roster = $derived(data.artists);
	const wikiPreview = $derived(data.articles);
	const marqueeItems = $derived(data.marqueeItems ?? []);

	function articleExcerpt(article: Article | WikiArticle): string {
		return article.excerpt ?? ('body' in article ? article.body.slice(0, 120) : '');
	}

	function articleCategory(article: Article | WikiArticle): string {
		return article.category ?? 'General';
	}

	function artistIsFeatured(artist: ArtistProfile): boolean {
		return Boolean(artist.featured);
	}
</script>

<svelte:head>
	<title>Mäkdäs — Modern Ethiopian art, held in common</title>
	<meta
		name="description"
		content="A community-held archive of modern Ethiopian art and its diaspora — artists, works, and the long conversation that connects them."
	/>
</svelte:head>

{#if featuredArtist}
	<FeaturedArtistHero artist={featuredArtist} posts={data.featuredPosts ?? []} />
{:else}
	<HomeFallbackHero />
{/if}
<MarqueeStrip items={marqueeItems} />

<section class="relative mx-auto max-w-[1600px] px-6 py-14 md:px-10 md:py-20">
	<SectionEyebrow number="02" label="From the archive" />

	<div class="mt-6 flex flex-wrap items-start justify-between gap-x-8 gap-y-4">
		<div class="max-w-2xl">
			<h2 class="font-display text-[2rem] leading-tight text-foreground md:text-[2.75rem]">
				Recent acquisitions
			</h2>
			<p class="mt-3 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-[15px]">
				Recently added to the collection. Works spanning abstraction, mixed media, and contemporary
				Ethiopian painting.
			</p>
		</div>
		<CtaLink href="/archive" variant="tertiary" class="mt-1 shrink-0 self-start">All works →</CtaLink>
	</div>

	{#if acquisitions.length > 0}
		<RecentAcquisitionsGrid items={acquisitions} />
	{:else}
		<EmptySectionPrompt
			class="mt-10"
			eyebrow="Growing collection"
			title="The archive is just beginning"
			body="Be among the first to add works to the community-held collection. Claim your portfolio and share what you're making."
			ctaLabel="Build your portfolio"
			ctaHref="/portfolio"
		/>
	{/if}
</section>

<section class="border-y border-border/60 bg-card/40">
	<div class="mx-auto max-w-[1600px] px-6 py-14 md:px-10 md:py-20">
		<div class="flex items-end justify-between gap-6">
			<div>
				<SectionEyebrow number="03" label="The keepers" />
				<h2 class="mt-3 font-display text-3xl text-foreground md:text-5xl">Artists in residence</h2>
			</div>
			<a
				href="/artists"
				class="hidden font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent sm:inline"
			>
				Full roster →
			</a>
		</div>

		{#if roster.length > 0}
			<div class="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
				{#each roster as a, i}
					{@const location = artistLocation(a)}
					{@const discipline = artistDiscipline(a) ?? ''}
					{@const badges = (() => {
						const b: { label: string; tone: 'accent' | 'muted' | 'secondary' }[] = [];
						if (artistIsFeatured(a)) b.push({ label: '● Featured', tone: 'accent' });
						if (
							location.includes('/') ||
							(location && !location.includes('Addis') && !location.includes('Bahir'))
						) {
							b.push({ label: '◐ Abroad', tone: 'secondary' });
						}
						if (i === roster.length - 1) b.push({ label: 'New', tone: 'accent' });
						if (discipline) b.push({ label: `✓ ${discipline.split(' / ')[0]}`, tone: 'muted' });
						return b;
					})()}
					<a href="/artists/{a.slug}" class="group block">
						<div class="relative aspect-[4/5] overflow-hidden rounded-sm bg-muted">
							{#if artistPortrait(a)}
								<img
									src={artistPortrait(a)}
									alt={artistName(a)}
									loading="lazy"
									class="h-full w-full object-cover grayscale transition duration-700 group-hover:scale-[1.03] group-hover:grayscale-0"
								/>
							{/if}
							<div
								class="absolute inset-x-0 bottom-0 flex flex-wrap gap-1.5 bg-gradient-to-t from-ink/80 to-transparent p-3"
							>
								{#each badges.slice(0, 2) as b}
									<span
										class="rounded-full px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.2em] backdrop-blur {b.tone ===
										'accent'
											? 'bg-accent/90 text-accent-foreground'
											: b.tone === 'secondary'
												? 'bg-secondary/85 text-secondary-foreground'
												: 'bg-background/80 text-foreground'}"
									>
										{b.label}
									</span>
								{/each}
							</div>
						</div>
						<div class="mt-3 flex items-baseline justify-between gap-3">
							<p class="font-display text-lg text-foreground">{artistName(a)}</p>
							{#if location}
								<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
									{location.split(' / ')[0]}
								</span>
							{/if}
						</div>
						{#if discipline}
							<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
								{discipline}
							</p>
						{/if}
					</a>
				{/each}
			</div>
		{:else}
			<EmptySectionPrompt
				class="mt-10"
				eyebrow="Open roster"
				title="Be among the first residents"
				body="The roster is forming. Claim your shareable profile and join the community archive as an early artist."
				ctaLabel="Claim your portfolio"
				ctaHref="/portfolio"
			/>
		{/if}
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
			Are you an artist?
			<span class="text-muted-foreground">Claim a shareable profile with Chapa & Telebirr checkout.</span>
		</p>
		<a
			href="/portfolio"
			class="font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent"
		>
			Portfolio builder →
		</a>
	</div>
</section>
