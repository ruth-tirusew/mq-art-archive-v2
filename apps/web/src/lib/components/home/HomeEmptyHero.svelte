<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';
	import WatercolorCanvas from '$lib/components/home/WatercolorCanvas.svelte';

	const today = new Date().toLocaleDateString('en-GB', {
		day: '2-digit',
		month: 'short',
		year: 'numeric'
	});

	// Placeholder tiles for work that isn't here yet. Paper tones come from the theme; a few tiles
	// carry the name of a discipline in mono, as a hint of what will fill the grid.
	const COLUMNS = 8;
	const ROWS = 6;
	const tones = ['bg-background', 'bg-card', 'bg-muted', 'bg-card', 'bg-background'];
	const disciplines = new Map<number, string>([
		[3, 'Painting'],
		[10, 'Ceramics'],
		[14, 'Textile'],
		[21, 'Photography'],
		[26, 'Sculpture'],
		[33, 'Print'],
		[37, 'Drawing'],
		[44, 'Calligraphy']
	]);
	const tiles = Array.from({ length: COLUMNS * ROWS }, (_, i) => ({
		i,
		tone: tones[(i * 7 + Math.floor(i / COLUMNS) * 3) % tones.length],
		label: disciplines.get(i)
	}));

</script>

<section
	class="relative flex min-h-[100svh] flex-col justify-between overflow-hidden border-b border-border/60 bg-border/70"
	aria-labelledby="home-title"
>
	<!-- The grid of tiles: empty paper for now. -->
	<div class="absolute inset-0 grid grid-cols-4 grid-rows-6 gap-px md:grid-cols-8" aria-hidden="true">
		{#each tiles as tile (tile.i)}
			<div class="{tile.tone} flex items-end p-2 md:p-3 {tile.i % COLUMNS >= 4 ? 'max-md:hidden' : ''}">
				{#if tile.label}
					<span class="t-eyebrow text-muted-foreground/60">{tile.label}</span>
				{/if}
			</div>
		{/each}
	</div>

	<WatercolorCanvas />

	<div class="pointer-events-none relative z-20 mx-auto flex w-full max-w-[1600px] items-center justify-between px-6 pt-6 t-eyebrow text-muted-foreground md:px-10">
		<span class="flex items-center gap-3"><span class="text-accent">●</span> Editorial archive</span>
		<span>{today}</span>
	</div>

	<div class="pointer-events-none relative z-20 mx-auto flex w-full max-w-[1600px] flex-1 flex-col justify-center px-6 py-16 md:px-10 md:py-24">
		<p class="t-eyebrow text-accent">Opening soon</p>
		<h1 id="home-title" class="mt-6 max-w-4xl font-display text-[16vw] leading-[0.9] tracking-tight text-foreground md:text-[9vw]">
			Art<em class="italic">iv</em>
		</h1>
		<p class="mt-6 max-w-2xl font-display text-2xl leading-snug text-foreground/90 md:text-3xl">
			Modern Ethiopian art, held in common.
		</p>
		<p class="mt-4 max-w-xl text-sm leading-relaxed text-muted-foreground md:text-base">
			The first artists are joining now. This wall fills as their work arrives.
		</p>

		<div class="pointer-events-auto mt-10 flex flex-wrap items-center gap-4">
			<CtaLink href="/portfolio" variant="primary">Claim your portfolio →</CtaLink>
			<CtaLink href="/about" variant="secondary">About the archive</CtaLink>
		</div>
	</div>

	<p class="pointer-events-none relative z-20 mx-auto w-full max-w-[1600px] px-6 pb-6 t-eyebrow text-muted-foreground/70 motion-reduce:hidden md:px-10">
		Move across the wall. It's waiting for colour.
	</p>
</section>
