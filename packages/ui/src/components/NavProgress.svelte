<script lang="ts">
	import { navigating } from '$app/state';

	/** Thin top bar shown while a navigation's load functions are in flight. Mount once in the root layout. */
	let visible = $state(false);

	$effect(() => {
		if (!navigating.to) {
			visible = false;
			return;
		}
		// Skip the bar for fast navigations so it never flashes.
		const t = setTimeout(() => (visible = true), 150);
		return () => clearTimeout(t);
	});
</script>

{#if visible}
	<div class="fixed inset-x-0 top-0 z-[60] h-0.5 overflow-hidden bg-accent/15" role="progressbar" aria-label="Loading page">
		<div class="nav-progress h-full w-1/3 bg-accent"></div>
	</div>
{/if}

<style>
	.nav-progress {
		animation: nav-slide 1.1s ease-in-out infinite;
	}
	@keyframes nav-slide {
		from { transform: translateX(-100%); }
		to { transform: translateX(300%); }
	}
	@media (prefers-reduced-motion: reduce) {
		.nav-progress { animation: none; width: 100%; }
	}
</style>
