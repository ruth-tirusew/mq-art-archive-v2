<script lang="ts">
	import {
		isEventBookmarked,
		toggleEventBookmark
	} from '$lib/application/eventBookmarks';
	import EventDetailView from '$lib/components/events/EventDetailView.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const event = $derived(data.event);
	let bookmarkVersion = $state(0);
	const bookmarked = $derived.by(() => {
		void bookmarkVersion;
		return event ? isEventBookmarked(event.id) : false;
	});

	function handleBookmark(id: string) {
		toggleEventBookmark(id);
		bookmarkVersion += 1;
	}
</script>

<svelte:head>
	<title>{event?.title ?? 'Event unavailable'} — Artiv Events</title>
	<meta name="description" content={(event?.description ?? '').slice(0, 160)} />
</svelte:head>

{#if event}
	<EventDetailView {event} {bookmarked} ontogglebookmark={handleBookmark} />
{:else}
	<section class="mx-auto max-w-[900px] px-6 py-24 text-center md:px-10">
		<h1 class="font-display text-4xl text-foreground">Event unavailable</h1>
		<p class="mt-4 text-muted-foreground">We couldn't load this event from the API.</p>
		<button
			type="button"
			onclick={() => location.reload()}
			class="mt-6 rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background"
		>
			Retry
		</button>
	</section>
{/if}
