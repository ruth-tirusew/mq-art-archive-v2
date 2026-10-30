<script lang="ts">
  import { Input, ListFooter, createPager } from '$ui';
  import { PAGE_SIZE } from '$lib/utils/paging';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { invalidateAll } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { eventsService } from '$lib/application/events';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';

  let { data } = $props();

  const pager = createPager({
    initial: data.events,
    pageSize: PAGE_SIZE,
    fetchPage: (offset, limit) => eventsService.list(data.status === 'all' ? undefined : (data.status as never), { limit, offset })
  });
  // The status filter lives in the URL, so a new filter arrives as new loader data.
  $effect(() => pager.reset(data.events));

  let query = $state('');
  const visible = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return pager.items;
    return pager.items.filter((event) =>
      [event.title, event.venue, event.city, event.event_type].some((field) => (field ?? '').toString().toLowerCase().includes(q))
    );
  });

  let syncing = $state(false);
  let syncMessage = $state('');
  let syncError = $state('');

  function formatDate(value?: string) {
    if (!value) return '—';
    return new Date(value).toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  async function syncEvents() {
    syncing = true;
    syncMessage = '';
    syncError = '';
    try {
      const result = await eventsService.sync();
      syncMessage = `Synced ${result.upserted} event${result.upserted === 1 ? '' : 's'}.`;
      await invalidateAll();
    } catch (err) {
      syncError = err instanceof Error ? err.message : 'Sync failed';
    } finally {
      syncing = false;
    }
  }
</script>

<svelte:head>
  <title>Events — Artiv admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="Events"
      title="Events"
      description="Create, edit, and review art events before they appear on the public site."
    />
  </div>
  <div class="mt-1 flex shrink-0 flex-wrap gap-2">
    <Button variant="secondary" disabled={syncing} onclick={syncEvents}>
      {syncing ? 'Syncing…' : 'Sync now'}
    </Button>
    <Button href="/events/new">New event</Button>
  </div>
</div>

{#if syncMessage}
  <Alert class="mb-4">{syncMessage}</Alert>
{/if}
{#if syncError}
  <Alert variant="destructive" class="mb-4">{syncError}</Alert>
{/if}

{#if pager.items.length === 0}
  <EmptyState
    title="No events in this view"
    description="Run a sync to pull in events from your sources, or add one manually."
    actionLabel="New event"
    actionHref="/events/new"
  />
{:else}
  <div class="mb-3 max-w-sm">
    <Input type="search" bind:value={query} placeholder="Search by title, venue or city" aria-label="Search events" />
  </div>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead class="hidden sm:table-cell">Type</TableHead>
        <TableHead>Status</TableHead>
        <TableHead class="hidden sm:table-cell">City</TableHead>
        <TableHead>Starts</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each visible as event}
        <TableRow>
          <TableCell>
            <a class="font-medium text-foreground underline decoration-accent underline-offset-4 hover:text-accent" href="/events/{event.id}">
              {event.title}
            </a>
          </TableCell>
          <TableCell class="hidden sm:table-cell">
            <Badge variant="outline">{event.event_type}</Badge>
          </TableCell>
          <TableCell><StatusBadge status={event.status} /></TableCell>
          <TableCell class="text-muted-foreground hidden sm:table-cell">{event.city || '—'}</TableCell>
          <TableCell class="text-muted-foreground">{formatDate(event.starts_at)}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
  {#if visible.length === 0}
    <p class="mt-6 text-sm text-muted-foreground">No loaded events match “{query}”.</p>
  {/if}
  <ListFooter
    shown={visible.length}
    loaded={pager.items.length}
    hasMore={pager.hasMore}
    loading={pager.loading}
    error={pager.error}
    searching={query.trim() !== ''}
    noun="events"
    onmore={() => pager.loadMore()}
  />
{/if}
