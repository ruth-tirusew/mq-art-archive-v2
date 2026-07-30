<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import MediaUploader from '$lib/components/MediaUploader.svelte';
  import { artistsService } from '$lib/application/artists';
  import type { ProfileStatus } from '$lib/core/domain/artist';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let email = $state('');
  let displayName = $state('');
  let handle = $state('');
  let slug = $state('');
  let discipline = $state('');
  let tagline = $state('');
  let bio = $state('');
  let portraitUrl = $state('');
  let location = $state('');
  let instagram = $state('');
  let status = $state<ProfileStatus>('draft');
  let saving = $state(false);
  let errorMessage = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    try {
      const artist = await artistsService.create({
        email: email.trim(),
        display_name: displayName.trim(),
        handle: handle.trim() || undefined,
        slug: slug.trim() || undefined,
        discipline,
        tagline,
        bio,
        portrait_url: portraitUrl,
        contact: { location, email: email.trim() },
        social: { instagram },
        status
      });
      await goto(`/artists/${artist.id}`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Create failed';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>New artist — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="04 — Artists"
  title="New artist"
  description="Create a user account and artist profile."
/>

<Card>
  <CardContent class="pt-6">
    <form class="space-y-5" onsubmit={handleSubmit}>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="email">Email</Label>
          <input id="email" type="email" required bind:value={email} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="display_name">Display name</Label>
          <input id="display_name" required bind:value={displayName} class={inputClass} />
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="handle">Handle</Label>
          <input id="handle" bind:value={handle} class={inputClass} placeholder="auto from name" />
        </div>
        <div class="space-y-2">
          <Label for="slug">Slug</Label>
          <input id="slug" bind:value={slug} class={inputClass} placeholder="auto from name" />
        </div>
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="discipline">Discipline</Label>
          <input id="discipline" bind:value={discipline} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="status">Status</Label>
          <select id="status" bind:value={status} class={inputClass}>
            <option value="draft">draft</option>
            <option value="pending">pending</option>
            <option value="approved">approved</option>
          </select>
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

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="location">Location</Label>
          <input id="location" bind:value={location} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="instagram">Instagram</Label>
          <input id="instagram" bind:value={instagram} class={inputClass} />
        </div>
      </div>

      {#if errorMessage}
        <Alert variant="destructive">{errorMessage}</Alert>
      {/if}

      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Create artist'}</Button>
        <Button type="button" variant="ghost" onclick={() => goto('/artists')}>Cancel</Button>
      </div>
    </form>
  </CardContent>
</Card>
