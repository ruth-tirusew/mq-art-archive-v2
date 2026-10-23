<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { onMount } from 'svelte';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import type { WikiSubmission } from '$lib/core/domain/wikiSubmission';
	import { ApiError } from '$lib/adapters/api/client';
	import { Alert, Badge, Button, EmptyState, ErrorState, PageHeader } from '$ui';

	let submissions = $state<WikiSubmission[]>([]);
	let toReview = $state<WikiSubmission[]>([]);
	let loading = $state(true);
	let error = $state('');
	let reviewQueueFailed = $state(false);

	async function load() {
		loading = true;
		error = '';
		reviewQueueFailed = false;
		try {
			submissions = await wikiSubmissionsService.listMine();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load wiki submissions';
		} finally {
			loading = false;
		}
		// Contributors have no articles to own, so this 403s harmlessly for them — only
		// artists ever have anything here. Any other failure must be visible: it hides review work.
		try {
			toReview = await wikiSubmissionsService.listOwnPending();
		} catch (e) {
			toReview = [];
			if (!(e instanceof ApiError && e.status === 403)) reviewQueueFailed = true;
		}
	}

	onMount(load);
</script>

<svelte:head><title>Wiki submissions — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<PageHeader eyebrow="Community wiki">
		Your <em class="italic">submissions</em>
		{#snippet actions()}<Button href="/studio/wiki/new">New submission</Button>{/snippet}
	</PageHeader>

	{#if reviewQueueFailed}
		<Alert tone="warning" title="Couldn’t check your review queue" class="mt-8">
			<div class="flex flex-wrap items-center gap-4">
				<span>Edits proposed to your articles may be waiting.</span>
				<Button variant="secondary" size="sm" onclick={load}>Try again</Button>
			</div>
		</Alert>
	{/if}

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
							<Badge tone="accent">Review</Badge>
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
		<div class="mt-8 space-y-3" role="status" aria-live="polite">
			<span class="sr-only">Loading submissions</span>
			<Skeleton class="h-16 w-full" />
			<Skeleton class="h-16 w-full" />
		</div>
	{:else if error}
		<div class="mt-8"><ErrorState message={error} onretry={load} /></div>
	{:else if submissions.length === 0}
		<EmptyState class="mt-8" title="You haven’t submitted anything yet." body="Propose a new entry or an edit. An editor reviews it before it goes live.">
			{#snippet action()}<Button variant="secondary" size="sm" href="/studio/wiki/new">Submit an article →</Button>{/snippet}
		</EmptyState>
	{:else}
		<div class="mt-4 divide-y divide-border/60 border-y border-border/60">
			{#each submissions as submission}
				<article class="py-5" data-testid="web-wiki-submission-item">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<h2 class="font-display text-xl text-foreground">{submission.title}</h2>
						<Badge tone={submission.status === 'approved' ? 'success' : submission.status === 'rejected' ? 'destructive' : 'warning'}>{submission.status}</Badge>
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
