<script lang="ts">
	import '../app.css';
	import SearchProvider from '$lib/components/SearchProvider.svelte';
	import SiteHeader from '$lib/components/SiteHeader.svelte';
	import SiteFooter from '$lib/components/SiteFooter.svelte';
	import { page } from '$app/stores';

	let { children }: { children: import('svelte').Snippet } = $props();

	const isHandlePage = $derived($page.url.pathname.startsWith('/@'));
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link
		href="https://fonts.googleapis.com/css2?family=Fraunces:ital,opsz,wght@0,9..144,300..900;1,9..144,300..900&family=Inter:wght@300..700&family=JetBrains+Mono:wght@400;500&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

<SearchProvider>
	<div class="page-enter min-h-screen bg-background">
		{#if !isHandlePage}
			<SiteHeader />
		{/if}
		{@render children()}
		{#if !isHandlePage}
			<SiteFooter />
		{/if}
	</div>
</SearchProvider>
