<script lang="ts">
  import { untrack } from 'svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { Badge } from '$lib/components/ui/badge';
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';
  import { Button } from '$ui';
  import { Alert } from '$lib/components/ui/alert';
  import { onboardingService } from '$lib/composition';
  import {
    APPLICATION_STATUS_FILTERS,
    type ApplicationStatusFilter,
    type OnboardingApplication
  } from '$lib/core/domain/onboarding';
  import { cn } from '$lib/utils';

  let { data } = $props();

  // The first page comes from the route loader (so it follows ?status=); "Load more" appends.
  let extra = $state<OnboardingApplication[]>([]);
  let loadingMore = $state(false);
  let loadError = $state('');

  // A different filter is a new load, so anything appended for the old one is dropped.
  $effect(() => {
    data.status;
    untrack(() => {
      extra = [];
      loadError = '';
    });
  });

  const applications = $derived([...data.applications, ...extra]);
  const hasMore = $derived(applications.length < data.total);
  const showingQueue = $derived(data.status === 'pending');

  const tabLabel: Record<ApplicationStatusFilter, string> = {
    pending: 'Pending',
    approved: 'Approved',
    rejected: 'Rejected',
    all: 'All'
  };
  const emptyCopy: Record<ApplicationStatusFilter, { title: string; description: string }> = {
    pending: { title: "You're all caught up", description: 'New artist applications will show up here for review.' },
    approved: { title: 'Nothing approved yet', description: 'Approved applications will be listed here.' },
    rejected: { title: 'Nothing rejected', description: 'Rejected applications will be listed here.' },
    all: { title: 'No applications yet', description: 'Applications will show up here as artists apply.' }
  };

  async function loadMore() {
    loadingMore = true;
    loadError = '';
    try {
      const page = await onboardingService.list(data.status, applications.length);
      extra = [...extra, ...page.items];
    } catch (cause) {
      loadError = cause instanceof Error ? cause.message : 'Could not load more applications';
    } finally {
      loadingMore = false;
    }
  }

  function formatDate(value?: string | null) {
    if (!value) return '—';
    return new Date(value).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>{tabLabel[data.status]} applications — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Applications"
  title={showingQueue ? 'Pending applications' : `${tabLabel[data.status]} applications`}
  description="Review artist applications. Approving one publishes the artist: their profile goes live with the work they submitted."
/>

<div class="mb-4 flex gap-2 overflow-x-auto" role="group" aria-label="Filter by status">
  {#each APPLICATION_STATUS_FILTERS as option}
    <a
      href={option === 'pending' ? '/applications' : `/applications?status=${option}`}
      aria-current={data.status === option ? 'page' : undefined}
      class={cn(
        'shrink-0 rounded-md border px-3 py-1 text-xs transition-colors',
        data.status === option
          ? 'border-foreground bg-muted text-foreground'
          : 'border-border text-muted-foreground hover:text-foreground'
      )}
    >
      {tabLabel[option]}
    </a>
  {/each}
</div>

{#if applications.length === 0}
  <EmptyState title={emptyCopy[data.status].title} description={emptyCopy[data.status].description} />
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Name</TableHead>
        <TableHead>Requested handle</TableHead>
        <TableHead>Type</TableHead>
        <TableHead>Submitted</TableHead>
        {#if !showingQueue}
          <TableHead>Decision</TableHead>
          <TableHead>Decided</TableHead>
        {/if}
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each applications as application}
        <TableRow data-testid="admin-application-row">
          <TableCell>
            <a class="font-medium text-foreground underline decoration-accent underline-offset-4 hover:text-accent" href="/applications/{application.id}">
              {application.display_name}
            </a>
          </TableCell>
          <TableCell class="font-mono text-xs text-muted-foreground">
            {application.requested_handle ? `@${application.requested_handle}` : '—'}
          </TableCell>
          <TableCell>
            <Badge variant="secondary">{application.applicant_type}</Badge>
          </TableCell>
          <TableCell class="text-muted-foreground">{formatDate(application.created_at)}</TableCell>
          {#if !showingQueue}
            <TableCell><StatusBadge status={application.status} /></TableCell>
            <TableCell class="text-muted-foreground">{formatDate(application.reviewed_at)}</TableCell>
          {/if}
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}

{#if loadError}
  <Alert variant="destructive" class="mt-4">{loadError}</Alert>
{/if}
{#if applications.length > 0}
  <p class="mt-3 text-xs text-muted-foreground">Showing {applications.length} of {data.total}.</p>
  {#if hasMore}
    <div class="mt-3">
      <Button variant="secondary" size="sm" loading={loadingMore} onclick={loadMore}>Load more</Button>
    </div>
  {/if}
{/if}
