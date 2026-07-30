<script lang="ts">
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
  <title>Events — mq admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="03 — Events"
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

{#if data.events.length === 0}
  <p class="text-sm text-muted-foreground">No events in this view.</p>
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead>Type</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>City</TableHead>
        <TableHead>Starts</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.events as event}
        <TableRow>
          <TableCell>
            <a class="font-medium text-foreground underline decoration-accent underline-offset-4 hover:text-accent" href="/events/{event.id}">
              {event.title}
            </a>
          </TableCell>
          <TableCell>
            <Badge variant="outline">{event.event_type}</Badge>
          </TableCell>
          <TableCell><Badge variant="secondary">{event.status}</Badge></TableCell>
          <TableCell class="text-muted-foreground">{event.city || '—'}</TableCell>
          <TableCell class="text-muted-foreground">{formatDate(event.starts_at)}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
