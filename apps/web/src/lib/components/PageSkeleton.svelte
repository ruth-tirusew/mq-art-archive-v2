<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';

	// A generic page-shaped placeholder: eyebrow, title, then `rows` content blocks.
	let {
		rows = 3,
		label = 'Loading',
		class: className = 'mx-auto max-w-2xl px-6 py-16 md:px-10 md:py-24'
	}: { rows?: number; label?: string; class?: string } = $props();
</script>

<div class={className} role="status" aria-live="polite">
	<span class="sr-only">{label}</span>
	<Skeleton class="h-3 w-24" />
	<Skeleton class="mt-5 h-10 w-2/3" />
	<div class="mt-10 space-y-4">
		{#each Array(rows) as _}
			<Skeleton class="h-16 w-full" />
		{/each}
	</div>
</div>
