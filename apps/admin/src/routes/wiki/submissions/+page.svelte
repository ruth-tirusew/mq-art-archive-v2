<script lang="ts">
  import { ConfirmDialog, createConfirm, toast } from '$ui';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { onMount } from 'svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import DiffView from '$lib/components/wiki/DiffView.svelte';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
  import type { WikiSubmission, WikiSubmissionReview } from '$lib/core/domain/wikiSubmission';

  const confirm = createConfirm();
  let submissions = $state<WikiSubmission[]>([]);
  let notes = $state<Record<string, string>>({});
  let busyId = $state('');
  let loading = $state(true);
  let error = $state('');

  let reviews = $state<Record<string, WikiSubmissionReview>>({});
  let expanded = $state<Record<string, boolean>>({});
  let diffLoading = $state<Record<string, boolean>>({});

  onMount(load);

  async function load() {
    error = '';
    loading = true;
    try {
      submissions = await wikiSubmissionsService.listPending();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not load submissions';
    } finally {
      loading = false;
    }
  }

  async function toggleDiff(item: WikiSubmission) {
    expanded[item.id] = !expanded[item.id];
    if (expanded[item.id] && !reviews[item.id]) {
      diffLoading[item.id] = true;
      try {
        reviews[item.id] = await wikiSubmissionsService.getForReview(item.id);
      } catch (e) {
        error = e instanceof Error ? e.message : 'Could not load diff';
      } finally {
        diffLoading[item.id] = false;
      }
    }
  }

  function askReview(item: WikiSubmission, action: 'approve' | 'reject') {
    confirm.ask(
      action === 'approve'
        ? {
            title: 'Approve this submission?',
            description: 'It is published to the public wiki immediately.',
            confirmLabel: 'Approve',
            run: () => review(item, action)
          }
        : {
            title: 'Reject this submission?',
            description: notes[item.id]?.trim()
              ? 'The submitter sees your note.'
              : 'You have not added a note, so the submitter will not know why.',
            confirmLabel: 'Reject',
            destructive: true,
            run: () => review(item, action)
          }
    );
  }

  async function review(item: WikiSubmission, action: 'approve' | 'reject') {
    busyId = item.id;
    error = '';
    try {
      if (action === 'approve') await wikiSubmissionsService.approve(item.id, notes[item.id]);
      else await wikiSubmissionsService.reject(item.id, notes[item.id]);
      submissions = submissions.filter((submission) => submission.id !== item.id);
      toast.success(action === 'approve' ? 'Submission approved and published.' : 'Submission rejected.');
    } catch (e) {
      error = e instanceof Error ? e.message : 'Review failed';
    } finally {
      busyId = '';
    }
  }
</script>

<svelte:head><title>Wiki submissions — Artiv admin</title></svelte:head>

<PageHeader
  eyebrow="Wiki"
  title="Community submissions"
  description="Review proposed wiki articles and edits."
/>

{#if error}
  <div class="mb-5 flex items-center justify-between gap-4 rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive" role="alert">
    <span>{error}</span>
    <Button variant="outline" size="sm" onclick={load}>Try again</Button>
  </div>
{/if}
{#if loading}
  <div class="space-y-3" role="status"><span class="sr-only">Loading submissions</span><Skeleton class="h-24 w-full" /><Skeleton class="h-24 w-full" /></div>
{:else if submissions.length === 0 && !error}
  <EmptyState title="No pending submissions" description="Community-proposed articles and edits will appear here." />
{:else}
  <div class="space-y-5">
    {#each submissions as item}
      {@const review_ = reviews[item.id]}
      {@const stale =
        review_?.article && item.based_on_version != null && review_.article.version !== item.based_on_version}
      <article class="rounded-xl border border-border/70 bg-card p-6" data-testid="admin-wiki-submission-item">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-medium text-foreground">{item.title}</h2>
            <p class="mt-1 font-mono text-[10px] text-muted-foreground">
              {item.kind === 'edit' ? `Edit to ${item.article_id}` : 'New article'} · submitter {item.submitter_id}
            </p>
          </div>
          <StatusBadge status={item.status} />
        </div>

        {#if item.kind === 'edit'}
          <Button variant="outline" size="sm" class="mt-4" onclick={() => toggleDiff(item)}>
            {expanded[item.id] ? 'Hide diff' : 'Show diff vs. current article'}
          </Button>
          {#if expanded[item.id]}
            {#if diffLoading[item.id]}
              <div class="mt-3 space-y-2" role="status"><span class="sr-only">Loading diff</span><Skeleton class="h-16 w-full" /></div>
            {:else if review_?.article}
              {#if stale}
                <p class="mt-3 text-xs text-amber-600 dark:text-amber-400">
                  This article changed since the submission was written (submitted against v{item.based_on_version},
                  now at v{review_.article.version}) — review carefully before approving.
                </p>
              {/if}
              {#if review_.article.title !== item.title}
                <p class="mt-3 text-xs text-muted-foreground">
                  Title: <span class="line-through text-red-600 dark:text-red-400">{review_.article.title}</span>
                  → <span class="text-emerald-700 dark:text-emerald-400">{item.title}</span>
                </p>
              {/if}
              <div class="mt-3">
                <DiffView oldText={review_.article.body} newText={item.body} />
              </div>
            {/if}
          {/if}
        {:else}
          <div class="mt-5 max-h-72 overflow-y-auto whitespace-pre-wrap rounded-md bg-muted/40 p-4 text-sm leading-relaxed">
            {item.body}
          </div>
        {/if}

        <textarea
          class="mt-4 min-h-20 w-full rounded-sm border border-input bg-background px-3 py-2 text-sm"
          placeholder="Optional review notes"
          value={notes[item.id] ?? ''}
          oninput={(event) => notes[item.id] = event.currentTarget.value}
        ></textarea>
        <div class="mt-4 flex gap-3">
          <Button disabled={busyId === item.id} onclick={() => askReview(item, 'approve')}>Approve</Button>
          <Button variant="destructive" disabled={busyId === item.id} onclick={() => askReview(item, 'reject')}>Reject</Button>
        </div>
      </article>
    {/each}
  </div>
{/if}
<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
