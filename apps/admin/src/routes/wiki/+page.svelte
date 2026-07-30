<script lang="ts">
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
  <title>Wiki — mq admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="06 — Wiki"
      title="Wiki articles"
      description="Write and publish handbook entries for the public wiki."
    />
  </div>
  <Button href="/wiki/new" class="mt-1 shrink-0">New article</Button>
</div>

{#if data.articles.length === 0}
  <p class="text-sm text-muted-foreground">No articles in this view.</p>
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead>Category</TableHead>
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
          <TableCell class="text-muted-foreground">{article.category || '—'}</TableCell>
          <TableCell><Badge variant="secondary">{article.status}</Badge></TableCell>
          <TableCell class="text-muted-foreground">{formatDate(article.updated_at)}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
