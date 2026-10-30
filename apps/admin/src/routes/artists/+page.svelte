<script lang="ts">
  import { Input, ListFooter, createPager } from '$ui';
  import { PAGE_SIZE } from '$lib/utils/paging';
  import { artistsService } from '$lib/application/artists';
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
    initial: data.artists,
    pageSize: PAGE_SIZE,
    fetchPage: (offset, limit) => artistsService.list(data.status === 'all' ? undefined : (data.status as never), { limit, offset })
  });
  // The status filter lives in the URL, so a new filter arrives as new loader data.
  $effect(() => pager.reset(data.artists));

  let query = $state('');
  const visible = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return pager.items;
    return pager.items.filter((artist) =>
      [artist.display_name, artist.handle, artist.slug].some((field) => (field ?? '').toString().toLowerCase().includes(q))
    );
  });
</script>

<svelte:head>
  <title>Artists — Artiv admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="Artists"
      title="Artist profiles"
      description="Browse artist profiles, edit their details and manage featured roster placement."
    />
  </div>
  <Button href="/artists/new" class="mt-1 shrink-0">New artist</Button>
</div>

{#if pager.items.length === 0}
  <EmptyState
    title="No artists in this view"
    description="Try another filter, or add an artist profile manually."
    actionLabel="New artist"
    actionHref="/artists/new"
  />
{:else}
  <div class="mb-3 max-w-sm">
    <Input type="search" bind:value={query} placeholder="Search by name or handle" aria-label="Search artists" />
  </div>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Name</TableHead>
        <TableHead>Status</TableHead>
        <TableHead class="hidden sm:table-cell">Featured</TableHead>
        <TableHead>Handle</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each visible as artist}
        <TableRow data-testid="admin-artist-row">
          <TableCell>
            <a
              class="font-medium underline decoration-accent underline-offset-4 hover:text-accent"
              href="/artists/{artist.id}"
            >
              {artist.display_name}
            </a>
          </TableCell>
          <TableCell><StatusBadge status={artist.status} /></TableCell>
          <TableCell class="hidden sm:table-cell">{artist.featured ? 'Yes' : '—'}</TableCell>
          <TableCell class="text-muted-foreground">@{artist.handle ?? artist.slug}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
  {#if visible.length === 0}
    <p class="mt-6 text-sm text-muted-foreground">No loaded artists match “{query}”.</p>
  {/if}
  <ListFooter
    shown={visible.length}
    loaded={pager.items.length}
    hasMore={pager.hasMore}
    loading={pager.loading}
    error={pager.error}
    searching={query.trim() !== ''}
    noun="artists"
    onmore={() => pager.loadMore()}
  />
{/if}
