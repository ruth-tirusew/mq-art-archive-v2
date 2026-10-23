<script lang="ts">
  import EmptyState from '$lib/components/EmptyState.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { Badge } from '$lib/components/ui/badge';
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';

  let { data } = $props();

  function formatDate(value?: string) {
    if (!value) return '—';
    return new Date(value).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }
</script>

<svelte:head>
  <title>Pending applications — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Applications"
  title="Pending applications"
  description="Review artist applications. Approving one creates a private draft profile the artist can publish themselves."
/>

{#if data.applications.length === 0}
  <EmptyState
    title="You're all caught up"
    description="New artist applications will show up here for review."
  />
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Name</TableHead>
        <TableHead>Requested handle</TableHead>
        <TableHead>Type</TableHead>
        <TableHead>Submitted</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.applications as application}
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
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
