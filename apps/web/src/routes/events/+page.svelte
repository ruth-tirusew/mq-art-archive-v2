<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import {
		listBookmarkedEventIds,
		toggleEventBookmark
	} from '$lib/application/eventBookmarks';
	import type { Event } from '$lib/core/domain/events';
	import EventsBrowseSidebar from '$lib/components/events/EventsBrowseSidebar.svelte';
	import EventsGridCard from '$lib/components/events/EventsGridCard.svelte';
	import EventsHero from '$lib/components/events/EventsHero.svelte';
	import EventsToolbar from '$lib/components/events/EventsToolbar.svelte';
	import EventSubmitForm from '$lib/components/events/EventSubmitForm.svelte';
	import EventTimelineCard from '$lib/components/events/EventTimelineCard.svelte';
	import {
		endOfWeek,
		groupEventsByDay,
		isSameDay,
		startOfWeek
	} from '$lib/components/events/eventFormat';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let rangeTab = $state<'upcoming' | 'past' | 'submissions'>('upcoming');
	let viewMode = $state<'list' | 'grid'>('list');
	let typeFilter = $state($page.url.searchParams.get('type') ?? 'All');
	let venueFilter = $state('All');
	let weekFilter = $state<'week' | 'all'>('all');
	let savedOnly = $state(false);
	let selectedDay = $state<Date | null>(null);
	let bookmarkVersion = $state(0);
	let localEvents = $state<Event[]>([]);
	let filtersReady = $state(false);

	$effect(() => {
		localEvents = [...data.events];
		filtersReady = true;
	});

	$effect(() => {
		void rangeTab;
		selectedDay = null;
	});

	$effect(() => {
		if (!filtersReady) return;
		const url = new URL($page.url);
		const current = url.searchParams.get('type') ?? 'All';
		if (typeFilter === current) return;
		if (typeFilter === 'All') url.searchParams.delete('type');
		else url.searchParams.set('type', typeFilter);
		void goto(`${url.pathname}${url.search}`, { replaceState: true, keepFocus: true, noScroll: true });
	});

	const bookmarkedIds = $derived.by(() => {
		void bookmarkVersion;
		return new Set(listBookmarkedEventIds());
	});

	const rangeEvents = $derived.by(() => {
		const now = Date.now();
		return localEvents.filter((e) => {
			const t = new Date(e.starts_at).getTime();
			if (rangeTab === 'upcoming') return t >= now;
			if (rangeTab === 'past') return t < now;
			return true;
		});
	});

	const types = $derived(
		[...new Set(rangeEvents.map((e) => e.event_type).filter(Boolean))].sort()
	);

	const venues = $derived(
		[...new Set(rangeEvents.map((e) => e.venue).filter((v): v is string => Boolean(v)))].sort()
	);

	const filtered = $derived.by(() => {
		const weekStart = startOfWeek().getTime();
		const weekEnd = endOfWeek().getTime();

		return rangeEvents.filter((e) => {
			const t = new Date(e.starts_at).getTime();
			if (typeFilter !== 'All' && e.event_type !== typeFilter) return false;
			if (venueFilter !== 'All' && (e.venue ?? '') !== venueFilter) return false;
			if (weekFilter === 'week' && (t < weekStart || t >= weekEnd)) return false;
			if (selectedDay && !isSameDay(new Date(e.starts_at), selectedDay)) return false;
			if (savedOnly && !bookmarkedIds.has(e.id)) return false;
			return true;
		});
	});

	// Drop stale type/venue when switching Upcoming/Past
	$effect(() => {
		if (localEvents.length === 0) return;
		if (typeFilter !== 'All' && types.length > 0 && !types.includes(typeFilter)) {
			typeFilter = 'All';
		}
		if (venueFilter !== 'All' && venues.length > 0 && !venues.includes(venueFilter)) {
			venueFilter = 'All';
		}
	});

	const grouped = $derived(groupEventsByDay(filtered));

	function handleBookmark(id: string) {
		toggleEventBookmark(id);
		bookmarkVersion += 1;
	}

	function onSubmitted(event: Event) {
		localEvents = [event, ...localEvents];
		rangeTab = 'upcoming';
	}
</script>

<svelte:head>
	<title>Events — Artiv</title>
	<meta
		name="description"
		content="A live calendar of Ethiopian gallery openings, poetry nights, theatre and design events — delivered weekly via Telegram."
	/>
</svelte:head>

<EventsHero events={localEvents} />

<EventsToolbar
	bind:rangeTab
	bind:viewMode
	bind:typeFilter
	bind:venueFilter
	bind:weekFilter
	bind:savedOnly
	{types}
	{venues}
/>

<section class="mx-auto max-w-[1600px] px-6 py-10 md:px-10 md:py-14">
	{#if data.error && localEvents.length === 0}
		<div class="rounded-sm border border-border bg-card/40 px-6 py-12 text-center">
			<h2 class="font-display text-2xl text-foreground">Events temporarily unavailable</h2>
			<p class="mt-3 text-sm text-muted-foreground">We couldn't load the calendar from the API.</p>
			<button
				type="button"
				onclick={() => location.reload()}
				class="mt-6 rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background"
			>
				Retry
			</button>
		</div>
	{:else if rangeTab === 'submissions'}
		<div class="grid gap-10 lg:grid-cols-[16rem_1fr]">
			<div class="hidden lg:block">
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
					Community
				</p>
				<p class="mt-3 max-w-xs text-sm text-muted-foreground">
					Help keep the calendar alive. Submitted events are reviewed before they join the public
					feed.
				</p>
			</div>
			<EventSubmitForm onsubmitted={onSubmitted} />
		</div>
	{:else}
		<div class="grid gap-10 lg:grid-cols-[16rem_1fr] xl:grid-cols-[18rem_1fr]">
			<div class="lg:sticky lg:top-24 lg:self-start">
				<EventsBrowseSidebar
					events={rangeEvents}
					bind:typeFilter
					bind:selectedDay
					{types}
				/>
			</div>

			<div class="min-w-0">
				{#if filtered.length === 0}
					<p
						class="py-16 text-center font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground"
					>
						No events match this filter.
					</p>
				{:else if viewMode === 'grid'}
					<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
						{#each filtered as event (event.id)}
							<EventsGridCard
								{event}
								bookmarked={bookmarkedIds.has(event.id)}
								ontogglebookmark={handleBookmark}
							/>
						{/each}
					</div>
				{:else}
					<div class="relative">
						<!-- Rail: line centered through 22px dot column -->
						<div
							class="absolute bottom-4 left-[10px] top-3 hidden w-px bg-border md:block"
							aria-hidden="true"
						></div>
						<div class="space-y-10">
							{#each grouped as group (group.key)}
								<section>
									<div class="mb-4 flex items-center gap-3">
										<span
											class="relative z-10 hidden h-[22px] w-[22px] shrink-0 items-center justify-center md:flex"
											aria-hidden="true"
										>
											<span class="h-2.5 w-2.5 rounded-full bg-accent ring-[3px] ring-background"
											></span>
										</span>
										<h2
											class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground"
										>
											{group.label}
										</h2>
									</div>
									<ul class="space-y-3 md:pl-[34px]">
										{#each group.events as event (event.id)}
											<li>
												<EventTimelineCard
													{event}
													bookmarked={bookmarkedIds.has(event.id)}
													ontogglebookmark={handleBookmark}
												/>
											</li>
										{/each}
									</ul>
								</section>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</section>
