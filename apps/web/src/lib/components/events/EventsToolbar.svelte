<script lang="ts">
	export type RangeTab = 'upcoming' | 'past' | 'submissions';
	export type ViewMode = 'list' | 'grid';
	export type WeekFilter = 'week' | 'all';

	let {
		rangeTab = $bindable('upcoming'),
		viewMode = $bindable('list'),
		typeFilter = $bindable('All'),
		venueFilter = $bindable('All'),
		weekFilter = $bindable('all'),
		savedOnly = $bindable(false),
		types,
		venues
	}: {
		rangeTab?: RangeTab;
		viewMode?: ViewMode;
		typeFilter?: string;
		venueFilter?: string;
		weekFilter?: WeekFilter;
		savedOnly?: boolean;
		types: string[];
		venues: string[];
	} = $props();

	const selectClass =
		'rounded-full border border-border bg-background px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-foreground focus:border-foreground focus:outline-none';
</script>

<section class="border-b border-border/70">
	<div
		class="mx-auto flex max-w-[1600px] flex-col gap-4 px-6 py-4 md:flex-row md:flex-wrap md:items-center md:px-10"
	>
		<div class="flex flex-wrap items-center gap-2">
			<button
				type="button"
				onclick={() => (rangeTab = 'upcoming')}
				class="rounded-full px-4 py-2 font-mono text-[10px] uppercase tracking-[0.2em] transition {rangeTab ===
				'upcoming'
					? 'bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
			>
				Upcoming
			</button>
			<button
				type="button"
				onclick={() => (rangeTab = 'past')}
				class="rounded-full px-4 py-2 font-mono text-[10px] uppercase tracking-[0.2em] transition {rangeTab ===
				'past'
					? 'bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
			>
				Past events
			</button>
			<button
				type="button"
				onclick={() => (rangeTab = 'submissions')}
				class="rounded-full px-4 py-2 font-mono text-[10px] uppercase tracking-[0.2em] transition {rangeTab ===
				'submissions'
					? 'bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
			>
				Submissions
			</button>
			<button
				type="button"
				onclick={() => (savedOnly = !savedOnly)}
				class="rounded-full border px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.18em] transition {savedOnly
					? 'border-accent bg-accent/10 text-accent'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				Saved
			</button>
		</div>

		{#if rangeTab !== 'submissions'}
			<div class="flex flex-wrap items-center gap-2 md:ml-auto">
				<select class={selectClass} bind:value={typeFilter} aria-label="Filter by type">
					<option value="All">All types</option>
					{#each types as t}
						<option value={t}>{t}</option>
					{/each}
				</select>
				<select class={selectClass} bind:value={venueFilter} aria-label="Filter by venue">
					<option value="All">All venues</option>
					{#each venues as v}
						<option value={v}>{v}</option>
					{/each}
				</select>
				<select class={selectClass} bind:value={weekFilter} aria-label="Filter by range">
					<option value="all">All dates</option>
					<option value="week">This week</option>
				</select>

				<div class="ml-1 flex items-center gap-1 border-l border-border/70 pl-3">
					<button
						type="button"
						onclick={() => (viewMode = 'list')}
						class="rounded p-1.5 transition {viewMode === 'list'
							? 'bg-foreground text-background'
							: 'text-muted-foreground hover:text-foreground'}"
						aria-label="List view"
						aria-pressed={viewMode === 'list'}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
							<path d="M4 6h16v2H4V6zm0 5h16v2H4v-2zm0 5h16v2H4v-2z" />
						</svg>
					</button>
					<button
						type="button"
						onclick={() => (viewMode = 'grid')}
						class="rounded p-1.5 transition {viewMode === 'grid'
							? 'bg-foreground text-background'
							: 'text-muted-foreground hover:text-foreground'}"
						aria-label="Grid view"
						aria-pressed={viewMode === 'grid'}
					>
						<svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
							<path d="M4 4h7v7H4V4zm9 0h7v7h-7V4zM4 13h7v7H4v-7zm9 0h7v7h-7v-7z" />
						</svg>
					</button>
				</div>
			</div>
		{/if}
	</div>
</section>
