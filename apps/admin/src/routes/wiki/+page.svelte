<script lang="ts">
  import EmptyState from '$lib/components/EmptyState.svelte';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
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
  <title>Wiki — Artiv admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="Wiki"
      title="Wiki articles"
      description="Write and publish handbook entries for the public wiki."
    />
  </div>
  <Button href="/wiki/new" class="mt-1 shrink-0">New article</Button>
</div>

{#if data.articles.length === 0}
  <EmptyState
    title="No articles in this view"
    description="Write a handbook entry, or review community submissions."
    actionLabel="New article"
    actionHref="/wiki/new"
  />
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead class="hidden sm:table-cell">Category</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Updated</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.articles as article}
        <TableRow data-testid="admin-article-row">
          <TableCell>
            <a
              class="font-medium underline decoration-accent underline-offset-4 hover:text-accent"
              href="/wiki/{article.id}"
            >
              {article.title}
            </a>
            {#if article.verified}
              <Badge variant="outline" class="ml-2">Verified</Badge>
            {/if}
          </TableCell>
          <TableCell class="text-muted-foreground hidden sm:table-cell">{article.category || '—'}</TableCell>
          <TableCell><StatusBadge status={article.status} /></TableCell>
          <TableCell class="text-muted-foreground">{formatDate(article.updated_at)}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
