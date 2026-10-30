<script lang="ts">
  import { Input, ListFooter, createPager } from '$ui';
  import { PAGE_SIZE } from '$lib/utils/paging';
  import { postsService } from '$lib/application/posts';
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

  const pager = createPager({
    initial: data.posts,
    pageSize: PAGE_SIZE,
    fetchPage: (offset, limit) => postsService.list(data.status === 'all' ? undefined : (data.status as never), { limit, offset })
  });
  // The status filter lives in the URL, so a new filter arrives as new loader data.
  $effect(() => pager.reset(data.posts));

  let query = $state('');
  const visible = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return pager.items;
    return pager.items.filter((post) =>
      [post.title, post.artist_name, post.medium, post.city].some((field) => (field ?? '').toString().toLowerCase().includes(q))
    );
  });
</script>

<svelte:head>
  <title>Posts — Artiv admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="Posts"
      title="Art posts"
      description="Moderate drafts, publications, and featured acquisitions."
    />
  </div>
  <Button href="/posts/new" class="mt-1 shrink-0">New post</Button>
</div>

{#if pager.items.length === 0}
  <EmptyState
    title="No posts in this view"
    description="Posts appear here once artists publish work or you add one."
    actionLabel="New post"
    actionHref="/posts/new"
  />
{:else}
  <div class="mb-3 max-w-sm">
    <Input type="search" bind:value={query} placeholder="Search by title, artist or medium" aria-label="Search posts" />
  </div>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Title</TableHead>
        <TableHead>Artist</TableHead>
        <TableHead>Status</TableHead>
        <TableHead class="hidden sm:table-cell">Featured</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each visible as post}
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
          <TableCell><StatusBadge status={post.status} /></TableCell>
          <TableCell class="hidden sm:table-cell">{post.featured_acquisition ? 'Yes' : '—'}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
  {#if visible.length === 0}
    <p class="mt-6 text-sm text-muted-foreground">No loaded posts match “{query}”.</p>
  {/if}
  <ListFooter
    shown={visible.length}
    loaded={pager.items.length}
    hasMore={pager.hasMore}
    loading={pager.loading}
    error={pager.error}
    searching={query.trim() !== ''}
    noun="posts"
    onmore={() => pager.loadMore()}
  />
{/if}
