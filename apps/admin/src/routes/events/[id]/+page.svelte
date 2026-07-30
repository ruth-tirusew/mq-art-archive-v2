<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { eventsService } from '$lib/application/events';
  import type { EventStatus } from '$lib/core/domain/event';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let { data } = $props();
  let event = $state(data.event);
  let notes = $state('');
  let saving = $state(false);
  let errorMessage = $state('');
  let message = $state('');

  function toLocalInput(iso?: string | null) {
    if (!iso) return '';
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return '';
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
  }

  let title = $state(data.event.title);
  let description = $state(data.event.description ?? '');
  let eventType = $state(data.event.event_type);
  let venue = $state(data.event.venue ?? '');
  let city = $state(data.event.city ?? '');
  let sourceUrl = $state(data.event.source_url ?? '');
  let imageUrl = $state(data.event.image_url ?? '');
  let startsAt = $state(toLocalInput(data.event.starts_at));
  let endsAt = $state(toLocalInput(data.event.ends_at));
  let status = $state<EventStatus>(data.event.status);

  function applyEvent(e: typeof event) {
    event = e;
    title = e.title;
    description = e.description ?? '';
    eventType = e.event_type;
    venue = e.venue ?? '';
    city = e.city ?? '';
    sourceUrl = e.source_url ?? '';
    imageUrl = e.image_url ?? '';
    startsAt = toLocalInput(e.starts_at);
    endsAt = toLocalInput(e.ends_at);
    status = e.status;
  }

  async function saveContent(ev: Event) {
    ev.preventDefault();
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyEvent(
        await eventsService.update(event.id, {
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
        })
      );
      message = 'Event saved.';
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Save failed';
    } finally {
      saving = false;
    }
  }

  async function review(reviewStatus: 'approved' | 'rejected') {
    saving = true;
    errorMessage = '';
    message = '';
    try {
      applyEvent(await eventsService.review(event.id, reviewStatus, notes));
      message = `Marked ${reviewStatus}.`;
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Review failed';
    } finally {
      saving = false;
    }
  }

  async function remove() {
    if (!confirm('Delete this event? This cannot be undone.')) return;
    saving = true;
    errorMessage = '';
    try {
      await eventsService.delete(event.id);
      await goto('/events');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Delete failed';
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>{event.title} — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="03 — Events"
  title={event.title}
  description="Edit event content and review publication status."
/>

<div class="space-y-6">
  <Card>
    <CardHeader>
      <div class="flex flex-wrap items-center gap-2">
        <Badge variant="secondary">{event.event_type}</Badge>
        <Badge>{event.status}</Badge>
      </div>
      <CardTitle class="pt-2">{event.title}</CardTitle>
    </CardHeader>
    <CardContent class="space-y-4">
      <div class="space-y-2">
        <Label for="notes">Review notes</Label>
        <Textarea id="notes" bind:value={notes} placeholder="Optional review notes" />
      </div>
      <div class="flex flex-wrap gap-2">
        <Button disabled={saving} onclick={() => review('approved')}>Approve</Button>
        <Button variant="destructive" disabled={saving} onclick={() => review('rejected')}>Reject</Button>
        <Button variant="outline" disabled={saving} onclick={() => remove()}>Delete</Button>
      </div>
      {#if message}<Alert>{message}</Alert>{/if}
      {#if errorMessage}<Alert variant="destructive">{errorMessage}</Alert>{/if}
      <Button variant="ghost" href="/events">Back to list</Button>
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
          <input id="source_url" bind:value={sourceUrl} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="image_url">Image URL</Label>
          <input id="image_url" bind:value={imageUrl} class={inputClass} />
        </div>
        <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save content'}</Button>
      </form>
    </CardContent>
  </Card>
</div>
