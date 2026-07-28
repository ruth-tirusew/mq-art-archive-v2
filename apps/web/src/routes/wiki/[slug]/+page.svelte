<script lang="ts">
	import { onMount } from 'svelte';
	import VerifiedChip from '$lib/components/VerifiedChip.svelte';
	import type { PageData } from './$types';
	import { recordPageView } from '$lib/application/analytics';

	let { data }: { data: PageData } = $props();

	const article = $derived(data.article);
	const title = $derived(article?.title ?? 'Wiki unavailable');
	onMount(() => {
		if (article?.id) recordPageView('article', article.id);
	});
</script>

<svelte:head>
	<title>{title} — Artiv Wiki</title>
	<meta name="description" content={article?.excerpt ?? 'A community-written guide for Ethiopian creatives.'} />
</svelte:head>

{#if article}
<article class="mx-auto max-w-[900px] px-6 py-16 md:px-10 md:py-24">
	<a
		href="/wiki"
		class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground hover:text-foreground"
	>
		← Back to wiki
	</a>

	<div class="mt-8 flex flex-wrap items-center gap-3">
		<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
			{article.category ?? 'General'}
		</span>
		{#if article.verified}
			<VerifiedChip />
		{/if}
		{#if article.updated_at}
			<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				· last edited {new Date(article.updated_at).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })}
			</span>
		{/if}
	</div>

	<h1 class="mt-5 font-display text-4xl leading-tight text-foreground md:text-5xl">{title}</h1>

	{#if article.excerpt}
		<p class="mt-6 text-lg leading-relaxed text-muted-foreground">{article.excerpt}</p>
	{/if}

	<div class="mt-10 space-y-6 text-base leading-relaxed text-foreground/90 md:text-lg wiki-body">
		{#if article.body}
			{#each article.body.split('\n\n') as paragraph}
				{#if paragraph.trim()}
					<p>{paragraph}</p>
				{/if}
			{/each}
		{/if}
	</div>
</article>
{:else}
	<section class="mx-auto max-w-[900px] px-6 py-24 text-center md:px-10">
		<h1 class="font-display text-4xl text-foreground">Wiki entry unavailable</h1>
		<p class="mt-4 text-muted-foreground">We couldn't load this entry from the API.</p>
		<button
			type="button"
			onclick={() => location.reload()}
			class="mt-6 rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background"
		>
			Retry
		</button>
	</section>
{/if}

<style>
	.wiki-body :global(p) {
		margin: 0;
	}
</style>
