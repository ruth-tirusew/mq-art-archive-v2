<script lang="ts">
	import FeaturedArtistHero from '$lib/components/home/FeaturedArtistHero.svelte';
	import HomeFallbackHero from '$lib/components/home/HomeFallbackHero.svelte';
	import EditorialCanvas from '$lib/components/home/EditorialCanvas.svelte';
	import MarqueeStrip from '$lib/components/home/MarqueeStrip.svelte';
	import HeldInCommon from '$lib/components/home/HeldInCommon.svelte';
	import FeaturedWorks from '$lib/components/home/FeaturedWorks.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const heroSpreads = $derived(data.heroSpreads ?? []);
	const featuredArtist = $derived(data.featuredArtist);
	const marqueeItems = $derived(data.marqueeItems ?? []);
	const works = $derived(data.acquisitions ?? []);
</script>

<svelte:head>
	<title>Artiv — Discover Ethiopian artists</title>
	<meta
		name="description"
		content="Discover Ethiopian artists and explore their worlds — a living archive of modern Ethiopian art."
	/>
</svelte:head>

{#if heroSpreads.length > 0}
	<EditorialCanvas spreads={heroSpreads} />
{:else}
	<HomeFallbackHero />
{/if}

<MarqueeStrip items={marqueeItems} />

{#if featuredArtist}
	<FeaturedArtistHero artist={featuredArtist} posts={data.featuredPosts ?? []} />
{/if}

<FeaturedWorks {works} excludeArtistSlug={featuredArtist?.slug} />

<HeldInCommon articles={data.articles ?? []} events={data.upcomingEvents ?? []} />
