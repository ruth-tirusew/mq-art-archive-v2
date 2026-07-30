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
  <title>Artists — mq admin</title>
</svelte:head>

<div class="mb-2 flex flex-wrap items-start justify-between gap-4">
  <div class="min-w-0 flex-1">
    <PageHeader
      eyebrow="04 — Artists"
      title="Artist profiles"
      description="Approve pending profiles and manage featured roster placement."
    />
  </div>
  <Button href="/artists/new" class="mt-1 shrink-0">New artist</Button>
</div>

{#if data.artists.length === 0}
  <p class="text-sm text-muted-foreground">No artists in this view.</p>
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Name</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Featured</TableHead>
        <TableHead>Handle</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.artists as artist}
        <TableRow data-testid="admin-artist-row">
          <TableCell>
            <a
              class="font-medium underline decoration-accent underline-offset-4 hover:text-accent"
              href="/artists/{artist.id}"
            >
              {artist.display_name}
            </a>
          </TableCell>
          <TableCell><Badge variant="secondary">{artist.status}</Badge></TableCell>
          <TableCell>{artist.featured ? 'Yes' : '—'}</TableCell>
          <TableCell class="text-muted-foreground">@{artist.handle ?? artist.slug}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}
