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
	<title>{event.title} — Artiv Events</title>
	<meta name="description" content={(event.description ?? '').slice(0, 160)} />
</svelte:head>

{#if event}
	<EventDetailView {event} {bookmarked} ontogglebookmark={handleBookmark} />
{/if}
