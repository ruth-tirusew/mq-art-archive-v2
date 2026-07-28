<script lang="ts">
	import type { Event } from '$lib/core/domain/events';
	import { sourceLabel } from '$lib/core/domain/events';
	import EventMedia from './EventMedia.svelte';
	import { fmtEventDate, truncate } from './eventFormat';

	let {
		event,
		bookmarked = false,
		ontogglebookmark
	}: {
		event: Event;
		bookmarked?: boolean;
		ontogglebookmark?: (id: string) => void;
	} = $props();

	const d = $derived(fmtEventDate(event.starts_at));
	const place = $derived([event.venue, event.city].filter(Boolean).join(' · '));
	const href = $derived(`/events/${encodeURIComponent(event.slug)}`);
</script>

<article
	class="group relative flex flex-col overflow-hidden border border-border/70 bg-background/60 transition hover:border-border hover:bg-card/70"
>
	<a {href} class="absolute inset-0 z-0" aria-label={event.title}></a>
	<div class="pointer-events-none relative z-10">
		<EventMedia {event} class="aspect-[4/3] w-full object-cover" />
	</div>
	<div class="pointer-events-none relative z-10 flex flex-1 flex-col p-4">
		<div class="flex items-start justify-between gap-2">
			<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-accent">{event.event_type}</p>
			<button
				type="button"
				class="pointer-events-auto rounded p-1 text-muted-foreground transition hover:text-accent"
				aria-label={bookmarked ? 'Remove bookmark' : 'Bookmark event'}
				onclick={(e) => {
					e.preventDefault();
					e.stopPropagation();
					ontogglebookmark?.(event.id);
				}}
			>
				<svg
					class="h-4 w-4"
					viewBox="0 0 24 24"
					fill={bookmarked ? 'currentColor' : 'none'}
					stroke="currentColor"
					stroke-width="1.8"
					aria-hidden="true"
				>
					<path d="M6 3h12a1 1 0 0 1 1 1v17l-7-4-7 4V4a1 1 0 0 1 1-1Z" />
				</svg>
			</button>
		</div>
		<h3 class="mt-2 font-display text-xl leading-snug text-foreground group-hover:text-accent">
			{event.title}
		</h3>
		<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.16em] text-muted-foreground">
			{d.month} {d.day} · {d.time}
		</p>
		{#if place}
			<p class="mt-1 text-sm text-muted-foreground">{place}</p>
		{/if}
		{#if event.description}
			<p class="mt-2 text-sm text-muted-foreground">{truncate(event.description, 100)}</p>
		{/if}
		<p class="mt-auto pt-4 font-mono text-[9px] uppercase tracking-[0.16em] text-muted-foreground">
			{sourceLabel(event.source_url)}
		</p>
	</div>
</article>
