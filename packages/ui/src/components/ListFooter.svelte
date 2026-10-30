<script lang="ts">
	import Button from './Button.svelte';

	type Props = {
		/** Rows currently shown (after any search). */
		shown: number;
		/** Rows fetched so far. */
		loaded: number;
		hasMore: boolean;
		loading?: boolean;
		error?: string;
		/** True when a search is narrowing the list, so the footer can say what it covers. */
		searching?: boolean;
		noun?: string;
		onmore: () => void;
	};

	let { shown, loaded, hasMore, loading = false, error = '', searching = false, noun = 'rows', onmore }: Props = $props();
</script>

<div class="mt-4 flex flex-wrap items-center justify-between gap-3">
	<p class="t-meta text-muted-foreground" aria-live="polite">
		{#if searching}
			{shown} of {loaded} loaded {noun} match{hasMore ? ' (search covers loaded rows only)' : ''}
		{:else}
			{loaded} {noun} loaded{hasMore ? '' : ' (all)'}
		{/if}
	</p>
	<div class="flex items-center gap-3">
		{#if error}<span class="text-sm text-destructive" role="alert">{error}</span>{/if}
		{#if hasMore}
			<Button variant="secondary" size="sm" {loading} onclick={onmore}>
				{error ? 'Try again' : searching ? 'Load more to search further' : 'Load more'}
			</Button>
		{/if}
	</div>
</div>
