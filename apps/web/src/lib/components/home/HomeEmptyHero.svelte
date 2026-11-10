<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';
	import WatercolorCanvas from '$lib/components/home/WatercolorCanvas.svelte';

	// A wall of blank tiles, waiting for work: six across, four down. A few are washed with a
	// faint tint from the site palette (see packages/ui theme.css); the rest are bare paper.
	const COLUMNS = 6;
	const ROWS = 4;
	const tints = new Map<number, 'tan' | 'sage' | 'blush' | 'peach'>([
		[0, 'tan'],
		[3, 'sage'],
		[4, 'blush'],
		[7, 'peach'],
		[14, 'tan'],
		[15, 'blush'],
		[20, 'sage'],
		[21, 'tan']
	]);
	const tiles = Array.from({ length: COLUMNS * ROWS }, (_, i) => ({ i, tint: tints.get(i) }));
</script>

<section
	class="relative flex min-h-[calc(100svh-4.2rem)] flex-col justify-between overflow-hidden border-b border-border/60 bg-border/70"
	style="--pigment-red: var(--accent); --pigment-ochre: var(--ochre); --pigment-olive: var(--success); --pigment-indigo: var(--indigo-deep);"
	aria-labelledby="home-title"
>
	<div class="absolute inset-0 grid grid-cols-6 grid-rows-4 gap-px" aria-hidden="true">
		{#each tiles as tile (tile.i)}
			<div class="tile {tile.tint ?? ''}"></div>
		{/each}
	</div>

	<WatercolorCanvas />

	<div class="pointer-events-none relative z-20 flex items-start justify-between px-6 pt-8 t-eyebrow text-muted-foreground md:px-12">
		<span>01 — A beginning</span>
		<span class="text-right">Ethiopian art directory</span>
	</div>

	<div class="pointer-events-none relative z-20 mx-auto flex w-full max-w-5xl flex-col items-center px-6 py-16 text-center">
		<h1 id="home-title" class="font-display text-[13vw] leading-[1.02] tracking-tight text-foreground md:text-[6.4vw]">
			A new canvas for
			<em class="block text-accent italic">Ethiopian art.</em>
		</h1>
		<p class="mt-8 max-w-md text-lg leading-relaxed text-muted-foreground md:text-xl">
			For the artists, the stories, and the worlds still waiting to be discovered.
		</p>
		<div class="pointer-events-auto mt-10">
			<CtaLink href="/login?mode=signup" variant="tertiary">
				Join the community <span aria-hidden="true">→</span>
			</CtaLink>
		</div>
	</div>

	<div class="pointer-events-none relative z-20 flex items-end justify-between gap-6 px-6 pb-8 md:px-12">
		<p class="max-w-[16rem] font-display text-lg text-muted-foreground italic md:max-w-none md:text-2xl">
		</p>
		<p class="text-right t-eyebrow text-muted-foreground">Artiv · Est. 2026</p>
	</div>
</section>

<style>
	/* Paper, with a trace of colour mixed in from the theme, so the tints follow the palette. */
	.tile {
		background: var(--background);
	}
	.tile.tan {
		background: color-mix(in oklab, var(--ochre) 9%, var(--background));
	}
	.tile.sage {
		background: color-mix(in oklab, var(--success) 14%, var(--background));
	}
	.tile.blush {
		background: color-mix(in oklab, var(--accent) 16%, var(--background));
	}
	.tile.peach {
		background: color-mix(in oklab, var(--ochre) 22%, var(--background));
	}
</style>
