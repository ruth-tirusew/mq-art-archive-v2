<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import type { Event } from '$lib/core/domain/events';
	import type { Event as StaticEvent } from '$lib/data/events';
	import { TELEGRAM_DIGEST_URL } from '$lib/data/events';
	import { formatEventDate, formatEventTime } from '$lib/utils/display';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let typeFilter = $state<string>('All');

	const events = $derived(data.events);

	const filtered = $derived(
		typeFilter === 'All'
			? events
			: events.filter((e) => eventType(e) === typeFilter)
	);

	function eventType(e: Event | StaticEvent): string {
		return 'event_type' in e ? e.event_type : e.type;
	}

	function eventStartsAt(e: Event | StaticEvent): string {
		return 'starts_at' in e ? e.starts_at : e.date;
	}

	function eventCity(e: Event | StaticEvent): string {
		return e.city ?? '';
	}

	function eventVenue(e: Event | StaticEvent): string {
		return e.venue ?? '';
	}

	function eventSlug(e: Event | StaticEvent): string {
		return e.slug;
	}

	function fmtDate(iso: string) {
		const d = new Date(iso);
		return {
			day: d.getDate().toString().padStart(2, '0'),
			month: d.toLocaleString('en', { month: 'short' }).toUpperCase(),
			weekday: d.toLocaleString('en', { weekday: 'long' })
		};
	}

	const types = $derived(
		data.types.length ? data.types : [...new Set(events.map((e) => eventType(e)))].sort()
	);
</script>

<svelte:head>
	<title>Events — Mäkdäs</title>
	<meta
		name="description"
		content="A live calendar of Ethiopian gallery openings, poetry nights, theatre and design events — delivered weekly via Telegram."
	/>
</svelte:head>

<section class="border-b border-border/60">
	<div class="mx-auto grid max-w-[1600px] gap-10 px-6 py-16 md:grid-cols-12 md:px-10 md:py-20">
		<div class="md:col-span-7">
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
				⁂ &nbsp; Pillar 03 · The discovery hub
			</p>
			<h1 class="mt-4 font-display text-4xl leading-[1.05] text-foreground md:text-6xl">
				Every opening, reading and pop-up — <em class="italic">in one place</em>.
			</h1>
			<p class="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg">
				We don't wait for venues to post here. A scraper crawls Telegram channels, Facebook pages and
				gallery sites every morning, then de-duplicates and tags. Institutional partners publish
				directly.
			</p>
		</div>

		<div class="md:col-span-5">
			<div class="rounded-sm border border-foreground bg-foreground p-7 text-background">
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-background/60">
					✺ &nbsp; Distribution
				</p>
				<h2 class="mt-3 font-display text-2xl leading-tight md:text-3xl">
					The weekly digest, in your Telegram.
				</h2>
				<p class="mt-4 text-sm leading-relaxed text-background/75">
					Every Sunday, our bot sends a curated rundown of the week's events to your inbox. No app, no
					extra account — just where the community already is.
				</p>
				<div class="mt-6 flex items-center gap-3">
					<input
						placeholder="@your-telegram"
						class="flex-1 rounded-full border border-background/30 bg-transparent px-4 py-2.5 font-mono text-[11px] text-background placeholder:text-background/40 focus:border-background focus:outline-none"
						readonly
						aria-label="Telegram username"
					/>
					<a
						href={TELEGRAM_DIGEST_URL}
						target="_blank"
						rel="noopener noreferrer"
						class="rounded-full bg-background px-4 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-foreground transition hover:bg-accent hover:text-background"
					>
						Subscribe →
					</a>
				</div>
				<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-background/40">
					1,247 subscribers · last digest sent Sun, Jun 21
				</p>
			</div>
		</div>
	</div>
</section>

<section class="border-b border-border/60">
	<div
		class="mx-auto flex max-w-[1600px] flex-wrap items-center gap-3 px-6 py-5 font-mono text-[10px] uppercase tracking-[0.2em] md:px-10"
	>
		<button
				type="button"
				onclick={() => (typeFilter = 'All')}
				class="rounded-full border px-3 py-1.5 transition {typeFilter === 'All'
					? 'border-foreground bg-foreground text-background'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				All
			</button>
		{#each types as f}
			<button
				type="button"
				onclick={() => (typeFilter = f)}
				class="rounded-full border px-3 py-1.5 transition {typeFilter === f
					? 'border-foreground bg-foreground text-background'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				{f}
			</button>
		{/each}
		<span class="ml-auto text-muted-foreground">
			{filtered.length} upcoming · auto-synced 06:00 EAT
		</span>
	</div>
</section>

<section class="mx-auto max-w-[1600px] px-6 py-10 md:px-10 md:py-14">
	<ul class="divide-y divide-border/70 border-y border-border/70">
		{#each filtered as e}
			{@const d = fmtDate(eventStartsAt(e))}
			<li>
				<a
					href="/events/{eventSlug(e)}"
					class="group grid grid-cols-12 items-center gap-6 py-7 transition hover:bg-card/40"
				>
					<div class="col-span-3 md:col-span-2">
						<p class="font-display text-4xl leading-none text-foreground md:text-5xl">{d.day}</p>
						<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
							{d.month} · {d.weekday.slice(0, 3)}
						</p>
					</div>

					<div class="col-span-9 md:col-span-6">
						<p class="font-display text-xl text-foreground transition group-hover:text-accent md:text-2xl">
							{e.title}
						</p>
						<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
							{eventVenue(e)} · {eventCity(e)}
						</p>
					</div>

					<p class="col-span-6 hidden font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground md:col-span-2 md:block">
						{eventType(e)}
					</p>

					<div class="col-span-6 text-right md:col-span-2">
						{#if 'source' in e}
							<span
								class="inline-flex items-center rounded-full border px-2.5 py-0.5 font-mono text-[9px] uppercase tracking-[0.2em] border-border bg-background text-muted-foreground"
							>
								{e.source === 'Institutional' ? '✓ partner' : e.source === 'Scraped' ? '⟲ scraped' : '↑ submitted'}
							</span>
							{#if e.host}
								<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
									via {e.host}
								</p>
							{/if}
						{:else}
							<span
								class="inline-flex items-center rounded-full border border-accent/40 bg-accent/10 px-2.5 py-0.5 font-mono text-[9px] uppercase tracking-[0.2em] text-accent"
							>
								✓ verified
							</span>
						{/if}
					</div>
				</a>
			</li>
		{/each}
	</ul>

	{#if filtered.length === 0}
		<p class="py-16 text-center font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
			No events match this filter.
		</p>
	{/if}
</section>
