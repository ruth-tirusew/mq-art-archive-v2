<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import BotanicalMark from '$lib/components/BotanicalMark.svelte';
  import BotanicalAdornment from '$lib/components/BotanicalAdornment.svelte';
  import MediaUploader from '$lib/components/MediaUploader.svelte';
  import { artistsService } from '$lib/application/artists';
  import type { ProfileStatus } from '$lib/core/domain/artist';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let { data } = $props();
  let artist = $state(data.artist);
  let saving = $state(false);
  let errorMessage = $state('');
  let message = $state('');

  let displayName = $state(data.artist.display_name);
  let handle = $state(data.artist.handle ?? '');
  let slug = $state(data.artist.slug);
  let discipline = $state(data.artist.discipline ?? '');
  let tagline = $state(data.artist.tagline ?? '');
  let bio = $state(data.artist.bio ?? '');
  let portraitUrl = $state(data.artist.portrait_url ?? '');
  let location = $state(data.artist.contact?.location ?? '');
  let contactEmail = $state(data.artist.contact?.email ?? '');
  let instagram = $state(data.artist.social?.instagram ?? '');

  function applyArtist(a: typeof artist) {
    artist = a;
    displayName = a.display_name;
    handle = a.handle ?? '';
    slug = a.slug;
    discipline = a.discipline ?? '';
    tagline = a.tagline ?? '';
    bio = a.bio ?? '';
    portraitUrl = a.portrait_url ?? '';
    location = a.contact?.location ?? '';
    contactEmail = a.contact?.email ?? '';
    instagram = a.social?.instagram ?? '';
  }

  async function saveContent(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyArtist(
        await artistsService.update(artist.id, {
          display_name: displayName.trim(),
          handle: handle.trim() || undefined,
          slug: slug.trim() || undefined,
          discipline,
          tagline,
          bio,
          portrait_url: portraitUrl,
          contact: { location, email: contactEmail },
          social: { instagram }
        })
      );
      message = 'Profile saved.';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Save failed';
    } finally {
      saving = false;
    }
  }

  async function setStatus(status: ProfileStatus) {
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyArtist(await artistsService.patch(artist.id, { status }));
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
      applyArtist(await artistsService.patch(artist.id, { featured: !artist.featured }));
      message = artist.featured ? 'Featured.' : 'Unfeatured.';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Update failed';
    } finally {
      saving = false;
    }
  }

  async function remove() {
    if (!confirm('Delete this artist profile? Only works if they have no posts.')) return;
    saving = true;
    errorMessage = '';
    try {
      await artistsService.delete(artist.id);
      await goto('/artists');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Delete failed';
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>{artist.display_name} — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="04 — Artists"
  title={artist.display_name}
  description="Edit profile content and moderate status."
/>

<div class="space-y-6">
  <Card class="relative overflow-hidden">
    <BotanicalAdornment
      variant="peony"
      class="absolute -left-4 -top-6 w-32 opacity-[0.2] sm:w-40"
    />
    <BotanicalAdornment
      variant="grass"
      class="absolute -right-3 top-8 w-20 opacity-[0.14] -rotate-6"
    />
    <BotanicalAdornment
      variant="bouquet"
      class="absolute -bottom-8 -right-6 w-36 opacity-[0.16] sm:w-44"
    />

    <CardHeader class="relative">
      <div class="flex flex-wrap items-start gap-4">
        <div class="h-20 w-20 shrink-0 overflow-hidden rounded-md bg-muted">
          {#if artist.portrait_url}
            <img src={artist.portrait_url} alt="" class="h-full w-full object-cover" />
          {:else}
            <div class="flex h-full w-full items-center justify-center">
              <BotanicalMark class="h-10 w-9 text-foreground/25" />
            </div>
          {/if}
        </div>
        <div class="min-w-0 space-y-2">
          <div class="flex flex-wrap items-center gap-2">
            <Badge>{artist.status}</Badge>
            {#if artist.featured}<Badge variant="secondary">Featured</Badge>{/if}
          </div>
          <CardTitle>@{artist.handle ?? artist.slug}</CardTitle>
          {#if artist.discipline}
            <p class="text-xs text-muted-foreground">{artist.discipline}</p>
          {/if}
        </div>
      </div>
    </CardHeader>
    <CardContent class="relative space-y-4">
      <div class="flex flex-wrap gap-2">
        <Button disabled={saving} onclick={() => setStatus('approved')}>Approve</Button>
        <Button variant="secondary" disabled={saving} onclick={() => setStatus('pending')}>Pending</Button>
        <Button variant="secondary" disabled={saving} onclick={() => setStatus('draft')}>Draft</Button>
        <Button variant="outline" disabled={saving} onclick={() => toggleFeatured()}>
          {artist.featured ? 'Unfeature' : 'Feature'}
        </Button>
        <Button variant="destructive" disabled={saving} onclick={() => remove()}>Delete</Button>
      </div>

      {#if message}<Alert>{message}</Alert>{/if}
      {#if errorMessage}<Alert variant="destructive">{errorMessage}</Alert>{/if}

      <Button variant="ghost" onclick={() => goto('/artists')}>Back to list</Button>
    </CardContent>
  </Card>

  <Card>
    <CardHeader>
      <CardTitle class="text-base">Profile</CardTitle>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" onsubmit={saveContent}>
        <div class="grid gap-4 sm:grid-cols-2">
          <div class="space-y-2">
            <Label for="display_name">Display name</Label>
            <input id="display_name" required bind:value={displayName} class={inputClass} />
          </div>
          <div class="space-y-2">
            <Label for="discipline">Discipline</Label>
            <input id="discipline" bind:value={discipline} class={inputClass} />
          </div>
          <div class="space-y-2">
            <Label for="handle">Handle</Label>
            <input id="handle" bind:value={handle} class={inputClass} />
          </div>
          <div class="space-y-2">
            <Label for="slug">Slug</Label>
            <input id="slug" bind:value={slug} class={inputClass} />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="tagline">Tagline</Label>
          <input id="tagline" bind:value={tagline} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="bio">Bio</Label>
          <Textarea id="bio" rows={5} bind:value={bio} />
        </div>
        <div class="space-y-2">
          <MediaUploader onUploaded={(media) => (portraitUrl = media.secure_url)} />
          {#if portraitUrl}
            <Label for="portrait_url">Portrait URL</Label>
            <input id="portrait_url" value={portraitUrl} readonly class={inputClass} />
          {/if}
        </div>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="space-y-2">
            <Label for="location">Location</Label>
            <input id="location" bind:value={location} class={inputClass} />
          </div>
          <div class="space-y-2">
            <Label for="contact_email">Contact email</Label>
            <input id="contact_email" bind:value={contactEmail} class={inputClass} />
          </div>
          <div class="space-y-2">
            <Label for="instagram">Instagram</Label>
            <input id="instagram" bind:value={instagram} class={inputClass} />
          </div>
        </div>
        <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save profile'}</Button>
      </form>
    </CardContent>
  </Card>
</div>
