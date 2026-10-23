<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import DiffView from '$lib/components/wiki/DiffView.svelte';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import type { WikiSubmissionReview } from '$lib/core/domain/wikiSubmission';
	import { Alert, Button, ConfirmDialog, createConfirm, ErrorState, Field, PageHeader, Textarea, toast } from '$ui';

	const id = $derived($page.params.id ?? '');

	let loading = $state(true);
	let error = $state('');
	let review = $state<WikiSubmissionReview | null>(null);
	let notes = $state('');
	let acting = $state(false);
	const confirm = createConfirm();

	async function load() {
		loading = true;
		error = '';
		try {
			review = await wikiSubmissionsService.getForReview(id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load submission';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	const stale = $derived(
		review?.article != null &&
			review.submission.based_on_version != null &&
			review.article.version !== review.submission.based_on_version
	);

	function ask(action: 'approve' | 'reject') {
		confirm.ask(
			action === 'approve'
				? {
						title: 'Approve this submission?',
						description: 'The change is published to the wiki immediately.',
						confirmLabel: 'Approve',
						run: () => act(action)
					}
				: {
						title: 'Reject this submission?',
						description: notes.trim()
							? 'The submitter sees your note.'
							: 'You have not added a note, so the submitter will not know why.',
						confirmLabel: 'Reject',
						destructive: true,
						run: () => act(action)
					}
		);
	}

	async function act(action: 'approve' | 'reject') {
		if (!review) return;
		acting = true;
		error = '';
		try {
			if (action === 'approve') await wikiSubmissionsService.approve(review.submission.id, notes);
			else await wikiSubmissionsService.reject(review.submission.id, notes);
			toast.success(action === 'approve' ? 'Approved and published.' : 'Rejected.');
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
	<div class="mb-8"><Button variant="link" href="/studio/wiki">← Submissions</Button></div>

	{#if loading}
		<div class="space-y-4" role="status" aria-live="polite">
			<span class="sr-only">Loading submission</span>
			<Skeleton class="h-8 w-1/2" />
			<Skeleton class="h-40 w-full" />
		</div>
	{:else if error && !review}
		<ErrorState message={error} onretry={load} />
	{:else if review}
		<PageHeader
			eyebrow="Community wiki · {review.submission.kind === 'edit' ? 'Proposed edit' : 'New article'}"
			lead="Submitted {new Date(review.submission.created_at).toLocaleDateString()}"
		>
			{review.submission.title}
		</PageHeader>

		{#if stale}
			<Alert tone="warning" title="Article changed" class="mt-6">
				This article changed since the submission was written (submitted against v{review.submission
					.based_on_version}, now at v{review.article?.version}). Review carefully before approving.
			</Alert>
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
			<div class="mt-6 max-h-96 overflow-y-auto whitespace-pre-wrap rounded-sm border border-border/60 bg-muted/20 p-4 text-sm leading-relaxed">
				{review.submission.body}
			</div>
		{/if}

		{#if review.submission.status === 'pending'}
			<div class="mt-6">
				<Field label="Review notes" hint="Optional for approval. Say what to change when rejecting.">
					{#snippet children({ id, describedby })}
						<Textarea {id} aria-describedby={describedby} rows={4} bind:value={notes} />
					{/snippet}
				</Field>
			</div>
			{#if error}<Alert tone="error" class="mt-4">{error}</Alert>{/if}
			<div class="mt-4 flex gap-3">
				<Button disabled={acting} onclick={() => ask('approve')}>Approve</Button>
				<Button variant="destructive" disabled={acting} onclick={() => ask('reject')}>Reject</Button>
			</div>
		{:else}
			<p class="t-meta mt-6 text-muted-foreground">Already {review.submission.status}</p>
		{/if}
	{/if}
</section>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
