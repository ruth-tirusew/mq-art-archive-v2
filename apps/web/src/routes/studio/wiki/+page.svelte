<script lang="ts">
	import { onMount } from 'svelte';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import type { WikiSubmission } from '$lib/core/domain/wikiSubmission';

	let submissions = $state<WikiSubmission[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			submissions = await wikiSubmissionsService.listMine();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load wiki submissions';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head><title>Wiki submissions — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<div class="flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Community wiki</p>
			<h1 class="mt-4 font-display text-4xl text-foreground">Your submissions</h1>
		</div>
		<a href="/studio/wiki/new" class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background">
			New submission
		</a>
	</div>

	{#if loading}
		<p class="mt-8 text-muted-foreground">Loading submissions…</p>
	{:else if error}
		<p class="mt-8 text-sm text-destructive" role="alert">{error}</p>
	{:else if submissions.length === 0}
		<p class="mt-8 text-sm text-muted-foreground">You have not submitted a wiki article yet.</p>
	{:else}
		<div class="mt-8 divide-y divide-border/60 border-y border-border/60">
			{#each submissions as submission}
				<article class="py-5">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<h2 class="font-display text-xl text-foreground">{submission.title}</h2>
						<span class="rounded-full border border-border px-3 py-1 font-mono text-[9px] uppercase tracking-[0.2em]">
							{submission.status}
						</span>
					</div>
					<p class="mt-2 line-clamp-2 text-sm text-muted-foreground">{submission.body}</p>
					{#if submission.review_notes}
						<p class="mt-3 text-xs text-muted-foreground">Review notes: {submission.review_notes}</p>
					{/if}
					<p class="mt-3 font-mono text-[9px] uppercase tracking-[0.18em] text-muted-foreground">
						{submission.article_id ? 'Article edit' : 'New article'} · {new Date(submission.created_at).toLocaleDateString()}
					</p>
				</article>
			{/each}
		</div>
	{/if}
</section>
