<script lang="ts">
	import type { Event } from '$lib/core/domain/events';
	import { sourceLabel } from '$lib/core/domain/events';
	import EventMedia from './EventMedia.svelte';
	import { fmtEventDate } from './eventFormat';

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
</script>

<article class="mx-auto max-w-[1100px] px-6 py-14 md:px-10 md:py-20">
	<a
		href="/events"
		class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground transition hover:text-foreground"
	>
		← Back to events
	</a>

	<div class="mt-8 overflow-hidden rounded-sm">
		<EventMedia {event} class="aspect-[21/9] w-full object-cover" />
	</div>

	<div class="mt-8 flex flex-wrap items-center gap-3">
		<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">{event.event_type}</span>
		{#if event.status === 'pending'}
			<span
				class="rounded-full border border-accent/30 bg-accent/10 px-2.5 py-0.5 font-mono text-[9px] uppercase tracking-[0.16em] text-accent"
			>
				Pending review
			</span>
		{/if}
		<button
			type="button"
			class="ml-auto inline-flex items-center gap-2 rounded-full border border-border px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.16em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
			aria-pressed={bookmarked}
			onclick={() => ontogglebookmark?.(event.id)}
		>
			<svg
				class="h-3.5 w-3.5"
				viewBox="0 0 24 24"
				fill={bookmarked ? 'currentColor' : 'none'}
				stroke="currentColor"
				stroke-width="1.8"
				aria-hidden="true"
			>
				<path d="M6 3h12a1 1 0 0 1 1 1v17l-7-4-7 4V4a1 1 0 0 1 1-1Z" />
			</svg>
			{bookmarked ? 'Saved' : 'Save'}
		</button>
	</div>

	<div class="mt-6 grid gap-8 md:grid-cols-12">
		<div class="md:col-span-3">
			<p class="font-display text-6xl leading-none text-foreground">{d.day}</p>
			<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.25em] text-accent">{d.month}</p>
			<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				{d.weekdayLong}
			</p>
			<p class="mt-4 font-mono text-[11px] uppercase tracking-[0.16em] text-foreground">{d.time}</p>
		</div>

		<div class="md:col-span-9">
			<h1 class="font-display text-4xl leading-tight text-foreground md:text-5xl">{event.title}</h1>
			{#if place}
				<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
					{place}
				</p>
			{/if}
			<p class="mt-3 font-mono text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
				Scanned from {sourceLabel(event.source_url)}
			</p>
		</div>
	</div>

	<div class="mt-12 max-w-3xl space-y-4 text-base leading-relaxed text-foreground/90 md:text-lg">
		{#each (event.description ?? '').split('\n\n') as para}
			{#if para.trim()}<p>{para}</p>{/if}
		{/each}
	</div>

	<p class="mt-10 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{d.full}</p>
</article>
