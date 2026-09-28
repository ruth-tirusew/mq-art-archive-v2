<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import DiffView from '$lib/components/wiki/DiffView.svelte';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import type { WikiSubmissionReview } from '$lib/core/domain/wikiSubmission';

	const id = $derived($page.params.id ?? '');

	let loading = $state(true);
	let error = $state('');
	let review = $state<WikiSubmissionReview | null>(null);
	let notes = $state('');
	let acting = $state(false);

	onMount(async () => {
		try {
			review = await wikiSubmissionsService.getForReview(id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load submission';
		} finally {
			loading = false;
		}
	});

	const stale = $derived(
		review?.article != null &&
			review.submission.based_on_version != null &&
			review.article.version !== review.submission.based_on_version
	);

	async function act(action: 'approve' | 'reject') {
		if (!review) return;
		acting = true;
		error = '';
		try {
			if (action === 'approve') await wikiSubmissionsService.approve(review.submission.id, notes);
			else await wikiSubmissionsService.reject(review.submission.id, notes);
			await goto('/studio/wiki');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Review failed';
		} finally {
			acting = false;
		}
	}
</script>

<svelte:head><title>Review submission — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<a
		href="/studio/wiki"
		class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">← Submissions</a
	>
	<p class="mt-8 font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Community wiki</p>

	{#if loading}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error && !review}
		<p class="mt-6 text-sm text-destructive" role="alert">{error}</p>
	{:else if review}
		<h1 class="mt-4 font-display text-4xl text-foreground">{review.submission.title}</h1>
		<p class="mt-3 font-mono text-[9px] uppercase tracking-[0.18em] text-muted-foreground">
			{review.submission.kind === 'edit' ? 'Proposed edit' : 'New article'} · submitted {new Date(
				review.submission.created_at
			).toLocaleDateString()}
		</p>

		{#if stale}
			<p class="mt-4 text-sm text-amber-600 dark:text-amber-400">
				This article changed since the submission was written (submitted against v{review.submission
					.based_on_version}, now at v{review.article?.version}) — review carefully before approving.
			</p>
		{/if}

		{#if review.article}
			{#if review.article.title !== review.submission.title}
				<p class="mt-6 text-sm text-muted-foreground">
					Title:
					<span class="text-destructive line-through">{review.article.title}</span>
					→ <span class="text-foreground">{review.submission.title}</span>
				</p>
			{/if}
			<div class="mt-4">
				<DiffView oldText={review.article.body} newText={review.submission.body} />
			</div>
		{:else}
			<div
				class="mt-6 max-h-96 overflow-y-auto whitespace-pre-wrap rounded-sm border border-border/60 bg-muted/20 p-4 text-sm leading-relaxed"
			>
				{review.submission.body}
			</div>
		{/if}

		{#if review.submission.status === 'pending'}
			<textarea
				class="field mt-6 min-h-20"
				placeholder="Optional review notes"
				bind:value={notes}
			></textarea>
			{#if error}<p class="mt-3 text-sm text-destructive" role="alert">{error}</p>{/if}
			<div class="mt-4 flex gap-3">
				<button
					type="button"
					class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
					disabled={acting}
					onclick={() => void act('approve')}
				>
					Approve
				</button>
				<button
					type="button"
					class="rounded-sm border border-destructive/40 px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-destructive disabled:opacity-50"
					disabled={acting}
					onclick={() => void act('reject')}
				>
					Reject
				</button>
			</div>
		{:else}
			<p class="mt-6 font-mono text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
				Already {review.submission.status}
			</p>
		{/if}
	{/if}
</section>

<style>
	.field {
		width: 100%;
		border: 1px solid color-mix(in oklab, var(--border) 70%, transparent);
		background: color-mix(in oklab, var(--card) 30%, transparent);
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		color: var(--foreground);
	}
	.field:focus {
		outline: 2px solid color-mix(in oklab, var(--accent) 50%, transparent);
		outline-offset: 2px;
	}
</style>
