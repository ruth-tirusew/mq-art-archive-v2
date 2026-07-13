<script lang="ts">
	import type { PageData } from './$types';
	import type { Event as StaticEvent } from '$lib/data/events';
	import type { Event as ApiEvent } from '$lib/core/domain/events';

	let { data }: { data: PageData } = $props();

	const event = $derived(data.event);

	function isStaticEvent(e: ApiEvent | StaticEvent): e is StaticEvent {
		return 'date' in e;
	}

	function eventDate(e: ApiEvent | StaticEvent): string {
		return isStaticEvent(e) ? e.date : e.starts_at;
	}

	function eventType(e: ApiEvent | StaticEvent): string {
		return isStaticEvent(e) ? e.type : e.event_type;
	}

	function eventDescription(e: ApiEvent | StaticEvent): string {
		return e.description ?? '';
	}

	function fmtDate(iso: string) {
		const d = new Date(iso);
		return {
			day: d.getDate().toString().padStart(2, '0'),
			month: d.toLocaleString('en', { month: 'short' }).toUpperCase(),
			weekday: d.toLocaleString('en', { weekday: 'long' }),
			full: d.toLocaleDateString('en-GB', {
				weekday: 'long',
				day: 'numeric',
				month: 'long',
				year: 'numeric'
			})
		};
	}

	const d = $derived(fmtDate(eventDate(event)));
</script>

<svelte:head>
	<title>{event.title} — Mäkdäs Events</title>
	<meta name="description" content={eventDescription(event).slice(0, 160)} />
</svelte:head>

<article class="mx-auto max-w-[900px] px-6 py-16 md:px-10 md:py-24">
	<a
		href="/events"
		class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground hover:text-foreground"
	>
		← Back to events
	</a>

	<div class="mt-8 flex flex-wrap items-center gap-3">
		<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">{eventType(event)}</span>
		{#if isStaticEvent(event)}
			<span
				class="inline-flex items-center rounded-full border px-2.5 py-0.5 font-mono text-[9px] uppercase tracking-[0.2em] border-border bg-background text-muted-foreground"
			>
				{event.source === 'Institutional' ? '✓ partner' : event.source === 'Scraped' ? '⟲ scraped' : '↑ submitted'}
			</span>
		{/if}
	</div>

	<div class="mt-6 grid gap-8 md:grid-cols-12">
		<div class="md:col-span-3">
			<p class="font-display text-6xl leading-none text-foreground">{d.day}</p>
			<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.25em] text-accent">{d.month}</p>
			<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				{d.weekday}
			</p>
			{#if isStaticEvent(event) && event.time}
				<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-foreground">
					{event.time}
				</p>
			{/if}
		</div>

		<div class="md:col-span-9">
			<h1 class="font-display text-4xl leading-tight text-foreground md:text-5xl">{event.title}</h1>
			<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				{event.venue} · {event.city}
			</p>
			{#if isStaticEvent(event) && event.host}
				<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
					Hosted by {event.host}
				</p>
			{/if}
		</div>
	</div>

	<div class="mt-12 space-y-4 text-base leading-relaxed text-foreground/90 md:text-lg">
		{#each eventDescription(event).split('\n\n') as para}
			{#if para.trim()}<p>{para}</p>{/if}
		{/each}
	</div>

	{#if isStaticEvent(event) && event.relatedArtistSlugs?.length}
		<div class="mt-12">
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
				Related artists
			</p>
			<div class="mt-3 flex flex-wrap gap-2">
				{#each event.relatedArtistSlugs as slug}
					<a
						href="/artists/{slug}"
						class="rounded-full border border-border bg-card/50 px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.15em] text-foreground transition hover:border-foreground"
					>
						{slug}
					</a>
				{/each}
			</div>
		</div>
	{/if}

	{#if isStaticEvent(event) && event.externalUrl}
		<a
			href={event.externalUrl}
			target="_blank"
			rel="noopener noreferrer"
			class="mt-10 inline-flex items-center gap-2 rounded-full bg-foreground px-5 py-2.5 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:bg-accent"
		>
			Venue website →
		</a>
	{/if}

	<p class="mt-10 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{d.full}</p>
</article>
