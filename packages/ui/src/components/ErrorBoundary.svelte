<script lang="ts">
	import type { Snippet } from 'svelte';
	import ErrorState from './ErrorState.svelte';

	/**
	 * Catches errors thrown while rendering or updating its children so one broken
	 * component shows a retry card instead of a blank page. Route load errors are not
	 * caught here; those go to +error.svelte.
	 */
	let { children, onerror }: { children: Snippet; onerror?: (error: unknown) => void } = $props();
</script>

<svelte:boundary onerror={(e) => (onerror ? onerror(e) : console.error(e))}>
	{@render children()}

	{#snippet failed(_error, reset)}
		<div class="container-page py-16">
			<ErrorState message="Part of this page failed to display." onretry={reset} />
		</div>
	{/snippet}
</svelte:boundary>
