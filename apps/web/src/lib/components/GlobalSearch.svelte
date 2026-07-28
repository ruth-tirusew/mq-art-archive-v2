<script lang="ts">
	import { getSearchContext, searchShortcutLabel } from './SearchProvider.svelte';

	const ctx = getSearchContext();
</script>

<button
	type="button"
	data-testid="web-search-trigger"
	onclick={() => ctx.setOpen(true)}
	class="inline-flex items-center gap-2 rounded-full border border-border bg-muted/60 px-2.5 py-1.5 font-mono text-[10px] uppercase tracking-[0.18em] text-muted-foreground transition hover:border-foreground/20 hover:bg-muted hover:text-foreground sm:gap-2.5 sm:px-3.5"
	aria-label="Open search ({searchShortcutLabel()})"
>
	<span aria-hidden="true" class="text-[11px] leading-none">+</span>
	<span class="hidden sm:inline">Search</span>
	<kbd
		class="hidden rounded border border-border/80 bg-background px-1.5 py-0.5 text-[9px] text-foreground/70 sm:inline"
	>
		{searchShortcutLabel()}
	</kbd>
</button>
