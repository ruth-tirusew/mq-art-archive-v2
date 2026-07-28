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
	class="group relative grid gap-4 border border-border/70 bg-background/60 p-4 transition hover:border-border hover:bg-card/70 md:grid-cols-[7.5rem_1fr_auto] md:gap-5 md:p-5"
>
	<!-- Stretch link sits behind content; interactive bits re-enable pointer events -->
	<a {href} class="absolute inset-0 z-0" aria-label={event.title}></a>

	<div class="pointer-events-none relative z-10 flex gap-3 md:block">
		<div class="min-w-[4.5rem]">
			<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				{d.month} {d.day}
			</p>
			<p class="mt-0.5 font-mono text-[10px] uppercase tracking-[0.2em] text-accent">{d.weekday}</p>
		</div>
		<EventMedia
			{event}
			class="h-16 w-20 shrink-0 rounded-sm object-cover md:mt-3 md:h-[4.5rem] md:w-full"
		/>
	</div>

	<div class="pointer-events-none relative z-10 min-w-0">
		<div class="flex flex-wrap items-center gap-2">
			<p class="font-mono text-[10px] uppercase tracking-[0.22em] text-accent">{event.event_type}</p>
			{#if event.status === 'pending'}
				<span
					class="rounded-full border border-accent/30 bg-accent/10 px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.16em] text-accent"
				>
					Pending review
				</span>
			{/if}
		</div>
		<h3
			class="mt-2 font-display text-xl leading-snug text-foreground transition group-hover:text-accent md:text-2xl"
		>
			{event.title}
		</h3>
		{#if place}
			<p
				class="mt-2 flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-[0.16em] text-muted-foreground"
			>
				<svg
					class="h-3 w-3 shrink-0"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					aria-hidden="true"
				>
					<path d="M12 21s7-5.5 7-11a7 7 0 1 0-14 0c0 5.5 7 11 7 11Z" />
					<circle cx="12" cy="10" r="2.5" />
				</svg>
				{place}
			</p>
		{/if}
		{#if event.description}
			<p class="mt-2 max-w-xl text-sm leading-relaxed text-muted-foreground">
				{truncate(event.description, 140)}
			</p>
		{/if}
	</div>

	<div
		class="pointer-events-none relative z-10 flex flex-row items-start justify-between gap-4 md:flex-col md:items-end md:text-right"
	>
		<div>
			<p class="font-mono text-[11px] uppercase tracking-[0.16em] text-foreground">{d.time}</p>
			<button
				type="button"
				class="pointer-events-auto mt-2 inline-flex rounded p-1.5 text-muted-foreground transition hover:text-accent"
				aria-label={bookmarked ? 'Remove bookmark' : 'Bookmark event'}
				aria-pressed={bookmarked}
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
		<p class="font-mono text-[9px] uppercase tracking-[0.16em] text-muted-foreground">
			Scanned from {sourceLabel(event.source_url)}
		</p>
	</div>
</article>
