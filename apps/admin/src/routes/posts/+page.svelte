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
</script>

<svelte:head>
  <title>Posts — mq admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="05 — Posts"
      title="Art posts"
      description="Moderate drafts, publications, and featured acquisitions."
    />
  </div>
  <Button href="/posts/new" class="mt-1 shrink-0">New post</Button>
</div>

{#if data.posts.length === 0}
  <p class="text-sm text-muted-foreground">No posts in this view.</p>
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead>Artist</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Featured</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.posts as post}
        <TableRow data-testid="admin-post-row">
          <TableCell>
            <a
              class="font-medium underline decoration-accent underline-offset-4 hover:text-accent"
              href="/posts/{post.id}"
            >
              {post.title}
            </a>
          </TableCell>
          <TableCell class="text-muted-foreground">{post.artist_name ?? '—'}</TableCell>
          <TableCell><Badge variant="secondary">{post.status}</Badge></TableCell>
          <TableCell>{post.featured_acquisition ? 'Yes' : '—'}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
