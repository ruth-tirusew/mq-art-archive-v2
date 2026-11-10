<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import ArtistAnswers from '$lib/components/artist/ArtistAnswers.svelte';
	import CopyLinkButton from '$lib/components/CopyLinkButton.svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import SocialMeta from '$lib/components/SocialMeta.svelte';
	import { acquisitionImage, artistName, artistPortrait } from '$lib/utils/fields';
	import { HEADER_ANSWER_KEYS, summarizeArtist } from '$lib/utils/artistSummary';
	import { profileUrl, shareDescription } from '$lib/utils/share';
	import { recordPageView } from '$lib/application/analytics';

	let { data }: { data: PageData } = $props();

	const isDemo = $derived(data.handle === 'demo');
	const summary = $derived(summarizeArtist(data.artist, data.answers));
	const shareUrl = $derived(profileUrl($page.url.origin, data.handle));
	// This is the link artists send around, so the card leads with their face, not their work.
	const sharePicture = $derived(artistPortrait(data.artist) ?? data.posts.map(acquisitionImage).find(Boolean));
	onMount(() => recordPageView('artist', data.artist.id));
</script>

<svelte:head>
	<title>@{data.handle} — Artiv</title>
	<meta name="description" content={shareDescription(summary.tagline, summary.bio)} />
</svelte:head>

<SocialMeta
	title="{artistName(data.artist)} (@{data.handle}) — Artiv"
	description={shareDescription(summary.tagline, summary.bio)}
	url={shareUrl}
	image={sharePicture}
/>

<header class="border-b border-border/60">
	<div class="mx-auto flex max-w-lg items-center justify-between px-6 py-4">
		<a href="/" class="font-display text-lg text-foreground">Artiv</a>
		<div class="flex items-center gap-4">
			<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">@{data.handle}</span>
			{#if !isDemo}<CopyLinkButton url={shareUrl} />{/if}
		</div>
	</div>
</header>

<main class="mx-auto max-w-lg px-6 py-12">
	<ShareableProfile artist={data.artist} works={data.posts} answers={data.answers} demo={isDemo} showHeader={false} />
	<ArtistAnswers answers={data.answers.filter((a) => !HEADER_ANSWER_KEYS.test(a.key))} artistName={artistName(data.artist)} compact />
</main>

<footer class="border-t border-border/60 py-6 text-center">
	<a
		href="/portfolio"
		class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground hover:text-foreground"
	>
		Create your own profile →
	</a>
</footer>
