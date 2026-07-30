<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import BotanicalMark from '$lib/components/BotanicalMark.svelte';
  import BotanicalAdornment from '$lib/components/BotanicalAdornment.svelte';
  import MediaUploader from '$lib/components/MediaUploader.svelte';
  import { postsService } from '$lib/application/posts';
  import type { ArtStatus } from '$lib/core/domain/art';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let { data } = $props();
  let post = $state(data.post);
  let artist = $state(data.artist);
  let saving = $state(false);
  let errorMessage = $state('');
  let message = $state('');

  let title = $state(data.post.title);
  let description = $state(data.post.description ?? '');
  let medium = $state(data.post.medium ?? '');
  let year = $state(data.post.year != null ? String(data.post.year) : '');
  let dimensions = $state(data.post.dimensions ?? '');
  let city = $state(data.post.city ?? '');
  let style = $state(data.post.style ?? '');
  let paletteText = $state((data.post.palette ?? []).join(', '));
  let mediaUrls = $state<string[]>((data.post.media ?? []).map((m) => m.url));

  const coverUrl = $derived(post.media?.[0]?.url);

  function applyPost(p: typeof post) {
    post = p;
    title = p.title;
    description = p.description ?? '';
    medium = p.medium ?? '';
    year = p.year != null ? String(p.year) : '';
    dimensions = p.dimensions ?? '';
    city = p.city ?? '';
    style = p.style ?? '';
    paletteText = (p.palette ?? []).join(', ');
    mediaUrls = (p.media ?? []).map((m) => m.url);
  }

  async function saveContent(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    message = '';
    try {
      const yearNum = year.trim() ? Number(year) : null;
      applyPost(
        await postsService.update(post.id, {
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
          media_urls: mediaUrls
        })
      );
      message = 'Post saved.';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Save failed';
    } finally {
      saving = false;
    }
  }

  async function setStatus(status: ArtStatus) {
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyPost(await postsService.patch(post.id, { status }));
      message = `Status set to ${status}.`;
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Update failed';
    } finally {
      saving = false;
    }
  }

  async function toggleFeatured() {
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyPost(
        await postsService.patch(post.id, {
          featured_acquisition: !post.featured_acquisition
        })
      );
      message = post.featured_acquisition ? 'Marked featured acquisition.' : 'Unfeatured.';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Update failed';
    } finally {
      saving = false;
    }
  }

  async function remove() {
    if (!confirm('Delete this post? This cannot be undone.')) return;
    saving = true;
    errorMessage = '';
    try {
      await postsService.delete(post.id);
      await goto('/posts');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Delete failed';
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>{post.title} — mq admin</title>
</svelte:head>

<PageHeader eyebrow="05 — Posts" title={post.title} description="Edit content, moderate status, and manage acquisition flag." />

<div class="grid gap-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(0,1fr)]">
  <div class="space-y-6">
    <Card class="overflow-hidden">
      <div class="aspect-[4/3] bg-muted">
        {#if coverUrl}
          <img src={coverUrl} alt={post.title} class="h-full w-full object-cover" />
        {:else}
          <div class="flex h-full w-full items-center justify-center">
            <BotanicalMark class="h-16 w-14 text-foreground/20" />
          </div>
        {/if}
      </div>
      <CardHeader>
        <div class="flex flex-wrap items-center gap-2">
          <Badge>{post.status}</Badge>
          {#if post.featured_acquisition}<Badge variant="secondary">Featured acquisition</Badge>{/if}
        </div>
      </CardHeader>
      <CardContent class="space-y-4">
        <div class="flex flex-wrap gap-2">
          <Button disabled={saving} onclick={() => setStatus('published')}>Publish</Button>
          <Button variant="secondary" disabled={saving} onclick={() => setStatus('draft')}>Draft</Button>
          <Button variant="secondary" disabled={saving} onclick={() => setStatus('archived')}>Archive</Button>
          <Button variant="outline" disabled={saving} onclick={() => toggleFeatured()}>
            {post.featured_acquisition ? 'Unfeature' : 'Feature acquisition'}
          </Button>
          <Button variant="destructive" disabled={saving} onclick={() => remove()}>Delete</Button>
        </div>

        {#if message}<Alert>{message}</Alert>{/if}
        {#if errorMessage}<Alert variant="destructive">{errorMessage}</Alert>{/if}

        <Button variant="ghost" onclick={() => goto('/posts')}>Back to list</Button>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle class="text-base">Content</CardTitle>
      </CardHeader>
      <CardContent>
        <form class="space-y-4" onsubmit={saveContent}>
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
            <input id="palette" bind:value={paletteText} class={inputClass} />
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
          <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save content'}</Button>
        </form>
      </CardContent>
    </Card>
  </div>

  <Card class="relative overflow-hidden self-start">
    <BotanicalAdornment
      variant="lily"
      class="absolute -left-3 -top-4 w-24 opacity-[0.18] sm:w-28"
    />
    <BotanicalAdornment
      variant="daisies"
      class="absolute -bottom-2 -right-3 w-28 opacity-[0.16] sm:w-32"
    />
    <BotanicalAdornment
      variant="slender"
      class="absolute -right-2 top-10 w-16 opacity-[0.12] rotate-12"
    />

    <CardHeader class="relative">
      <CardTitle class="text-base">Artist</CardTitle>
    </CardHeader>
    <CardContent class="relative">
      {#if artist}
        <a
          href="/artists/{artist.id}"
          class="group flex gap-4 rounded-lg border border-border/70 bg-card/80 p-4 backdrop-blur-[2px] transition-colors hover:border-accent/40 hover:bg-muted/40"
        >
          <div class="h-16 w-16 shrink-0 overflow-hidden rounded-md bg-muted">
            {#if artist.portrait_url}
              <img
                src={artist.portrait_url}
                alt=""
                class="h-full w-full object-cover"
              />
            {:else}
              <div class="flex h-full w-full items-center justify-center">
                <BotanicalMark class="h-8 w-7 text-foreground/25" />
              </div>
            {/if}
          </div>
          <div class="min-w-0 space-y-1">
            <p class="truncate font-medium text-foreground group-hover:text-accent">
              {artist.display_name}
            </p>
            <p class="truncate text-xs text-muted-foreground">
              @{artist.handle ?? artist.slug}
              {#if artist.discipline}
                · {artist.discipline}
              {/if}
            </p>
            <p class="line-clamp-3 text-sm text-muted-foreground">
              {artist.tagline || artist.bio || 'No bio.'}
            </p>
            <div class="flex flex-wrap gap-1.5 pt-1">
              {#if artist.status}<Badge variant="secondary">{artist.status}</Badge>{/if}
              {#if artist.featured}<Badge variant="secondary">Featured</Badge>{/if}
            </div>
          </div>
        </a>
      {:else}
        <p class="text-sm text-muted-foreground">Artist profile unavailable.</p>
      {/if}
    </CardContent>
  </Card>
</div>
