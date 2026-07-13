<script lang="ts">
	import VerifiedChip from '$lib/components/VerifiedChip.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const article = $derived(data.article);
	const title = $derived(article.title);
	const updated = $derived(
		'updated_at' in article && article.updated_at
			? article.updated_at
			: 'updated' in article
				? article.updated
				: new Date().toISOString()
	);
	const body = $derived('body' in article ? article.body : '');
</script>

<svelte:head>
	<title>{title} — Mäkdäs Wiki</title>
	<meta name="description" content={article.excerpt ?? 'A community-written guide for Ethiopian creatives.'} />
</svelte:head>

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
		<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
			· last edited {new Date(updated).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })}
		</span>
	</div>

	<h1 class="mt-5 font-display text-4xl leading-tight text-foreground md:text-5xl">{title}</h1>

	{#if article.excerpt}
		<p class="mt-6 text-lg leading-relaxed text-muted-foreground">{article.excerpt}</p>
	{/if}

	<div class="mt-10 space-y-6 text-base leading-relaxed text-foreground/90 md:text-lg wiki-body">
		{#if body}
			{#each body.split('\n\n') as paragraph}
				{#if paragraph.trim()}
					<p>{paragraph}</p>
				{/if}
			{/each}
		{:else}
			<p>
				This is a community-written entry. The body of the article would live here — structured into
				sections with diagrams, tables of fees, downloadable form templates, and footnotes citing the
				relevant Ethiopian statute or office.
			</p>
		{/if}
	</div>
</article>

<style>
	.wiki-body :global(p) {
		margin: 0;
	}
</style>
