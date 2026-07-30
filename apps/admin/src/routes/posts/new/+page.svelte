<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import MediaUploader from '$lib/components/MediaUploader.svelte';
  import { postsService } from '$lib/application/posts';
  import type { ArtStatus } from '$lib/core/domain/art';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let { data } = $props();

  let artistId = $state(data.artists[0]?.id ?? '');
  let title = $state('');
  let description = $state('');
  let medium = $state('');
  let year = $state('');
  let dimensions = $state('');
  let city = $state('');
  let style = $state('');
  let paletteText = $state('');
  let mediaUrls = $state<string[]>([]);
  let status = $state<ArtStatus>('draft');
  let saving = $state(false);
  let errorMessage = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    if (!artistId) {
      errorMessage = 'Select an artist.';
      return;
    }
    saving = true;
    errorMessage = '';
    try {
      const yearNum = year.trim() ? Number(year) : null;
      const post = await postsService.create({
        artist_id: artistId,
        title: title.trim(),
        description,
        medium,
        year: yearNum != null && Number.isFinite(yearNum) ? yearNum : null,
        dimensions,
        city,
        style,
        palette: paletteText
          .split(',')
          .map((s) => s.trim())
          .filter(Boolean),
        media_urls: mediaUrls,
        status
      });
      await goto(`/posts/${post.id}`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Create failed';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>New post — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="05 — Posts"
  title="New post"
  description="Create an art post on behalf of an artist."
/>

<Card>
  <CardContent class="pt-6">
    <form class="space-y-5" onsubmit={handleSubmit}>
      <div class="space-y-2">
        <Label for="artist_id">Artist</Label>
        <select id="artist_id" required bind:value={artistId} class={inputClass}>
          {#if data.artists.length === 0}
            <option value="">No artists available</option>
          {:else}
            {#each data.artists as a}
              <option value={a.id}>{a.display_name}</option>
            {/each}
          {/if}
        </select>
      </div>

      <div class="space-y-2">
        <Label for="title">Title</Label>
        <input id="title" required bind:value={title} class={inputClass} />
      </div>

      <div class="space-y-2">
        <Label for="description">Description</Label>
        <Textarea id="description" rows={4} bind:value={description} />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="medium">Medium</Label>
          <input id="medium" bind:value={medium} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="year">Year</Label>
          <input id="year" inputmode="numeric" bind:value={year} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="dimensions">Dimensions</Label>
          <input id="dimensions" bind:value={dimensions} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="city">City</Label>
          <input id="city" bind:value={city} class={inputClass} />
        </div>
      </div>

      <div class="space-y-2">
        <Label for="style">Style</Label>
        <input id="style" bind:value={style} class={inputClass} />
      </div>

      <div class="space-y-2">
        <Label for="palette">Palette (comma-separated)</Label>
        <input id="palette" bind:value={paletteText} class={inputClass} placeholder="#2c1810, #c4a574" />
      </div>

      <div class="space-y-2">
        <MediaUploader
          onUploaded={(media) => {
            if (!mediaUrls.includes(media.secure_url)) mediaUrls = [...mediaUrls, media.secure_url];
          }}
        />
        {#if mediaUrls.length}
          <ul class="grid gap-3 sm:grid-cols-2">
            {#each mediaUrls as url, index}
              <li class="flex gap-3 rounded-md border p-2">
                <img src={url} alt="" class="h-16 w-16 shrink-0 rounded object-cover" />
                <div class="min-w-0 flex-1">
                  <p class="truncate text-xs text-muted-foreground">{url}</p>
                  <Button
                    type="button"
                    variant="ghost"
                    onclick={() => (mediaUrls = mediaUrls.filter((_, itemIndex) => itemIndex !== index))}
                  >
                    Remove
                  </Button>
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      <div class="space-y-2">
        <Label for="status">Status</Label>
        <select id="status" bind:value={status} class={inputClass}>
          <option value="draft">draft</option>
          <option value="published">published</option>
          <option value="archived">archived</option>
        </select>
      </div>

      {#if errorMessage}
        <Alert variant="destructive">{errorMessage}</Alert>
      {/if}

      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving || !artistId}>{saving ? 'Saving…' : 'Create post'}</Button>
        <Button type="button" variant="ghost" onclick={() => goto('/posts')}>Cancel</Button>
      </div>
    </form>
  </CardContent>
</Card>
