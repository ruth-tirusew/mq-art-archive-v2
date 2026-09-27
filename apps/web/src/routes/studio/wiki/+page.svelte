<script lang="ts">
	import { onMount } from 'svelte';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import type { WikiSubmission } from '$lib/core/domain/wikiSubmission';

	let submissions = $state<WikiSubmission[]>([]);
	let toReview = $state<WikiSubmission[]>([]);
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
		// Contributors have no articles to own, so this 403s harmlessly for them — only
		// artists ever have anything here.
		try {
			toReview = await wikiSubmissionsService.listOwnPending();
		} catch {
			toReview = [];
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

	{#if !loading && toReview.length > 0}
		<div class="mt-10">
			<h2 class="font-display text-xl text-foreground">Awaiting your review</h2>
			<p class="mt-1 text-sm text-muted-foreground">
				Proposed edits to articles you own — pointing readers to what changed.
			</p>
			<div class="mt-4 divide-y divide-border/60 border-y border-border/60">
				{#each toReview as submission}
					<a
						href={`/studio/wiki/${submission.id}`}
						data-testid="web-wiki-review-item"
						class="block py-4 transition hover:bg-muted/20"
					>
						<div class="flex flex-wrap items-center justify-between gap-3">
							<h3 class="font-display text-lg text-foreground">{submission.title}</h3>
							<span
								class="rounded-full border border-accent px-3 py-1 font-mono text-[9px] uppercase tracking-[0.2em] text-accent"
							>
								Review
							</span>
						</div>
						<p class="mt-2 line-clamp-2 text-sm text-muted-foreground">{submission.body}</p>
					</a>
				{/each}
			</div>
		</div>
	{/if}

	{#if toReview.length > 0}
		<h2 class="mt-10 font-display text-xl text-foreground">Submitted by you</h2>
	{/if}

	{#if loading}
		<p class="mt-8 text-muted-foreground">Loading submissions…</p>
	{:else if error}
		<p class="mt-8 text-sm text-destructive" role="alert">{error}</p>
	{:else if submissions.length === 0}
		<p class="mt-8 text-sm text-muted-foreground">You have not submitted a wiki article yet.</p>
	{:else}
		<div class="mt-4 divide-y divide-border/60 border-y border-border/60">
			{#each submissions as submission}
				<article class="py-5" data-testid="web-wiki-submission-item">
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
