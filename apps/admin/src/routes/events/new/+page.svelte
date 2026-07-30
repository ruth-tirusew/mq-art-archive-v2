<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { eventsService } from '$lib/application/events';
  import type { EventStatus } from '$lib/core/domain/event';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let title = $state('');
  let description = $state('');
  let eventType = $state('Opening');
  let venue = $state('');
  let city = $state('');
  let sourceUrl = $state('');
  let imageUrl = $state('');
  let startsAt = $state('');
  let endsAt = $state('');
  let status = $state<EventStatus>('pending');
  let saving = $state(false);
  let errorMessage = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    try {
      const created = await eventsService.create({
        title: title.trim(),
        description,
        event_type: eventType,
        venue,
        city,
        source_url: sourceUrl || undefined,
        image_url: imageUrl || null,
        starts_at: new Date(startsAt).toISOString(),
        ends_at: endsAt ? new Date(endsAt).toISOString() : null,
        status
      });
      await goto(`/events/${created.id}`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Create failed';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>New event — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="03 — Events"
  title="New event"
  description="Manually add an art event for review or publication."
/>

<Card>
  <CardContent class="pt-6">
    <form class="space-y-5" onsubmit={handleSubmit}>
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
          <Label for="event_type">Type</Label>
          <input id="event_type" bind:value={eventType} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="status">Status</Label>
          <select id="status" bind:value={status} class={inputClass}>
            <option value="pending">pending</option>
            <option value="approved">approved</option>
            <option value="rejected">rejected</option>
          </select>
        </div>
        <div class="space-y-2">
          <Label for="venue">Venue</Label>
          <input id="venue" bind:value={venue} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="city">City</Label>
          <input id="city" bind:value={city} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="starts_at">Starts</Label>
          <input id="starts_at" type="datetime-local" required bind:value={startsAt} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="ends_at">Ends</Label>
          <input id="ends_at" type="datetime-local" bind:value={endsAt} class={inputClass} />
        </div>
      </div>

      <div class="space-y-2">
        <Label for="source_url">Source URL</Label>
        <input id="source_url" bind:value={sourceUrl} class={inputClass} placeholder="Optional" />
      </div>

      <div class="space-y-2">
        <Label for="image_url">Image URL</Label>
        <input id="image_url" bind:value={imageUrl} class={inputClass} />
      </div>

      {#if errorMessage}
        <Alert variant="destructive">{errorMessage}</Alert>
      {/if}

      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Create event'}</Button>
        <Button type="button" variant="ghost" onclick={() => goto('/events')}>Cancel</Button>
      </div>
    </form>
  </CardContent>
</Card>
