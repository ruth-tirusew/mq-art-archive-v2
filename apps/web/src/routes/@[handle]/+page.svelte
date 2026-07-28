<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import { recordPageView } from '$lib/application/analytics';

	let { data }: { data: PageData } = $props();

	const isDemo = $derived(data.handle === 'demo');
	onMount(() => recordPageView('artist', data.artist.id));
</script>

<svelte:head>
	<title>@{data.handle} — Artiv</title>
</svelte:head>

<header class="border-b border-border/60">
	<div class="mx-auto flex max-w-lg items-center justify-between px-6 py-4">
		<a href="/" class="font-display text-lg text-foreground">Artiv</a>
		<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
			@{data.handle}
		</span>
	</div>
</header>

<main class="mx-auto max-w-lg px-6 py-12">
	{#if isDemo}
		<p
			class="mb-6 rounded-sm border border-accent/30 bg-accent/10 px-4 py-3 text-center font-mono text-[10px] uppercase tracking-[0.2em] text-accent"
		>
			Demo profile — live data from the archive API
		</p>
	{/if}
	<ShareableProfile artist={data.artist} works={data.posts} demo={isDemo} showHeader={false} />
</main>

<footer class="border-t border-border/60 py-6 text-center">
	<a
		href="/portfolio"
		class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground hover:text-foreground"
	>
		Create your own profile →
	</a>
</footer>
