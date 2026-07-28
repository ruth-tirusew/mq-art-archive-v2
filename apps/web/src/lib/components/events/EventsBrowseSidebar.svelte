<script lang="ts">
	import type { Event } from '$lib/core/domain/events';
	import { isSameDay } from './eventFormat';

	let {
		events,
		typeFilter = $bindable('All'),
		selectedDay = $bindable<Date | null>(null),
		types
	}: {
		events: Event[];
		typeFilter?: string;
		selectedDay?: Date | null;
		types: string[];
	} = $props();

	let viewMonth = $state(new Date());

	const typeCounts = $derived.by(() => {
		const counts = new Map<string, number>();
		for (const e of events) {
			counts.set(e.event_type, (counts.get(e.event_type) ?? 0) + 1);
		}
		return types.map((t) => ({ type: t, count: counts.get(t) ?? 0 }));
	});

	const eventDays = $derived.by(() => {
		const set = new Set<string>();
		for (const e of events) {
			const d = new Date(e.starts_at);
			set.add(`${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`);
		}
		return set;
	});

	const calendarCells = $derived.by(() => {
		const year = viewMonth.getFullYear();
		const month = viewMonth.getMonth();
		const first = new Date(year, month, 1);
		const startPad = (first.getDay() + 6) % 7; // Monday-first
		const daysInMonth = new Date(year, month + 1, 0).getDate();
		const cells: { date: Date | null; hasEvent: boolean; isToday: boolean; selected: boolean }[] =
			[];
		for (let i = 0; i < startPad; i++) cells.push({ date: null, hasEvent: false, isToday: false, selected: false });
		const today = new Date();
		for (let day = 1; day <= daysInMonth; day++) {
			const date = new Date(year, month, day);
			const key = `${year}-${month}-${day}`;
			cells.push({
				date,
				hasEvent: eventDays.has(key),
				isToday: isSameDay(date, today),
				selected: selectedDay ? isSameDay(date, selectedDay) : false
			});
		}
		return cells;
	});

	const monthLabel = $derived(
		viewMonth.toLocaleString('en', { month: 'long', year: 'numeric' })
	);

	function prevMonth() {
		viewMonth = new Date(viewMonth.getFullYear(), viewMonth.getMonth() - 1, 1);
	}
	function nextMonth() {
		viewMonth = new Date(viewMonth.getFullYear(), viewMonth.getMonth() + 1, 1);
	}
	function goToday() {
		const today = new Date();
		viewMonth = new Date(today.getFullYear(), today.getMonth(), 1);
		selectedDay = today;
	}
	function selectDay(date: Date) {
		if (selectedDay && isSameDay(selectedDay, date)) {
			selectedDay = null;
			return;
		}
		selectedDay = date;
	}
</script>

<aside class="space-y-8">
	<div>
		<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Browse</p>
		<ul class="mt-3 space-y-0.5">
			<li>
				<button
					type="button"
					onclick={() => (typeFilter = 'All')}
					class="flex w-full items-center justify-between rounded-sm px-3 py-2.5 text-left font-mono text-[11px] uppercase tracking-[0.16em] transition {typeFilter ===
					'All'
						? 'bg-card text-foreground'
						: 'text-muted-foreground hover:bg-card/60 hover:text-foreground'}"
				>
					<span>All events</span>
					<span>{events.length}</span>
				</button>
			</li>
			{#each typeCounts as item}
				<li>
					<button
						type="button"
						onclick={() => (typeFilter = item.type)}
						class="flex w-full items-center justify-between rounded-sm px-3 py-2.5 text-left font-mono text-[11px] uppercase tracking-[0.16em] transition {typeFilter ===
						item.type
							? 'bg-card text-foreground'
							: 'text-muted-foreground hover:bg-card/60 hover:text-foreground'}"
					>
						<span>{item.type}</span>
						<span>{item.count}</span>
					</button>
				</li>
			{/each}
		</ul>
	</div>

	<div class="border-t border-border/70 pt-6">
		<div class="flex items-center justify-between">
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Calendar</p>
			<div class="flex gap-1">
				<button
					type="button"
					onclick={prevMonth}
					class="rounded px-2 py-1 font-mono text-[10px] text-muted-foreground hover:text-foreground"
					aria-label="Previous month">‹</button
				>
				<button
					type="button"
					onclick={nextMonth}
					class="rounded px-2 py-1 font-mono text-[10px] text-muted-foreground hover:text-foreground"
					aria-label="Next month">›</button
				>
			</div>
		</div>
		<p class="mt-3 font-display text-lg text-foreground">{monthLabel}</p>
		<div class="mt-3 grid grid-cols-7 gap-1 text-center font-mono text-[9px] uppercase tracking-[0.12em] text-muted-foreground">
			{#each ['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'] as dow}
				<span>{dow}</span>
			{/each}
		</div>
		<div class="mt-1 grid grid-cols-7 gap-1">
			{#each calendarCells as cell}
				{#if cell.date}
					<button
						type="button"
						onclick={() => selectDay(cell.date!)}
						class="relative flex h-8 items-center justify-center rounded-full font-mono text-[11px] transition
							{cell.selected || cell.isToday
							? 'bg-foreground text-background'
							: 'text-foreground hover:bg-card'}
							{cell.hasEvent && !cell.selected && !cell.isToday ? 'font-semibold' : ''}"
					>
						{cell.date.getDate()}
						{#if cell.hasEvent}
							<span
								class="absolute bottom-0.5 h-1 w-1 rounded-full {cell.selected || cell.isToday
									? 'bg-background'
									: 'bg-accent'}"
							></span>
						{/if}
					</button>
				{:else}
					<span></span>
				{/if}
			{/each}
		</div>
		<div class="mt-4 flex gap-2">
			<button
				type="button"
				onclick={goToday}
				class="flex-1 rounded-full border border-border py-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground transition hover:border-foreground hover:text-foreground"
			>
				Today
			</button>
			{#if selectedDay}
				<button
					type="button"
					onclick={() => (selectedDay = null)}
					class="rounded-full border border-accent/40 bg-accent/10 px-3 py-2 font-mono text-[10px] uppercase tracking-[0.16em] text-accent"
				>
					Clear day
				</button>
			{/if}
		</div>
	</div>
</aside>
