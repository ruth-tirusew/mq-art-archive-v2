<script lang="ts">
  import { onMount } from 'svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
  import type { WikiSubmission } from '$lib/core/domain/wikiSubmission';

  let submissions = $state<WikiSubmission[]>([]);
  let notes = $state<Record<string, string>>({});
  let busyId = $state('');
  let loading = $state(true);
  let error = $state('');

  onMount(load);

  async function load() {
    try {
      submissions = await wikiSubmissionsService.listPending();
    } catch (e) {
      error = e instanceof Error ? e.message : 'Could not load submissions';
    } finally {
      loading = false;
    }
  }

  async function review(item: WikiSubmission, action: 'approve' | 'reject') {
    busyId = item.id;
    error = '';
    try {
      if (action === 'approve') await wikiSubmissionsService.approve(item.id, notes[item.id]);
      else await wikiSubmissionsService.reject(item.id, notes[item.id]);
      submissions = submissions.filter((submission) => submission.id !== item.id);
    } catch (e) {
      error = e instanceof Error ? e.message : 'Review failed';
    } finally {
      busyId = '';
    }
  }
</script>

<svelte:head><title>Wiki submissions — mq admin</title></svelte:head>

<PageHeader
  eyebrow="06 — Wiki"
  title="Community submissions"
  description="Review proposed wiki articles and edits."
/>

{#if error}<p class="mb-5 text-sm text-destructive" role="alert">{error}</p>{/if}
{#if loading}
  <p class="text-sm text-muted-foreground">Loading submissions…</p>
{:else if submissions.length === 0}
  <p class="text-sm text-muted-foreground">No pending wiki submissions.</p>
{:else}
  <div class="space-y-5">
    {#each submissions as item}
      <article class="rounded-xl border border-border/70 bg-card p-6">
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <h2 class="text-lg font-medium text-foreground">{item.title}</h2>
            <p class="mt-1 font-mono text-[10px] text-muted-foreground">
              {item.article_id ? `Edit to ${item.article_id}` : 'New article'} · submitter {item.submitter_id}
            </p>
          </div>
          <Badge variant="secondary">{item.status}</Badge>
        </div>
        <div class="mt-5 max-h-72 overflow-y-auto whitespace-pre-wrap rounded-md bg-muted/40 p-4 text-sm leading-relaxed">
          {item.body}
        </div>
        <textarea
          class="mt-4 min-h-20 w-full rounded-sm border border-input bg-background px-3 py-2 text-sm"
          placeholder="Optional review notes"
          value={notes[item.id] ?? ''}
          oninput={(event) => notes[item.id] = event.currentTarget.value}
        ></textarea>
        <div class="mt-4 flex gap-3">
          <Button disabled={busyId === item.id} onclick={() => review(item, 'approve')}>Approve</Button>
          <Button variant="destructive" disabled={busyId === item.id} onclick={() => review(item, 'reject')}>Reject</Button>
        </div>
      </article>
    {/each}
  </div>
{/if}
