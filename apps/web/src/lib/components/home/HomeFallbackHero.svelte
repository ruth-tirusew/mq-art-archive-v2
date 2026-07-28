<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';

	const today = new Date().toLocaleDateString('en-GB', {
		day: '2-digit',
		month: 'short',
		year: 'numeric'
	});
</script>

<section
	class="relative flex h-[100vh] min-h-[100vh] flex-col justify-between overflow-hidden border-b border-border/60 bg-ink"
>
	<div
		class="relative z-20 mx-auto flex w-full max-w-[1600px] items-center justify-between px-6 pt-6 font-mono text-[10px] uppercase tracking-[0.3em] text-cream/70 md:px-10"
	>
		<span class="flex items-center gap-3">
			<span class="text-accent">●</span> Editorial archive
		</span>
		<span>{today}</span>
	</div>

	<div class="relative mx-auto flex w-full max-w-[1600px] flex-1 flex-col justify-center px-6 py-16 md:px-10 md:py-24">
		<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">02 / Welcome</p>
		<h1 class="mt-6 max-w-4xl font-display text-[14vw] leading-[0.9] tracking-tight text-cream md:text-[7.5vw]">
			Artiv
		</h1>
		<p class="mt-6 max-w-2xl font-display text-2xl leading-snug text-cream/90 md:text-3xl">
			Modern Ethiopian art, held in common.
		</p>
		<p class="mt-4 max-w-xl text-sm leading-relaxed text-cream/70 md:text-base">
			Artists, works, and the long conversation that connects them — held in common by the
			community.
		</p>

		<div class="mt-10 flex flex-wrap items-center gap-4">
			<CtaLink href="/portfolio" variant="primary" tone="on-dark">Claim your portfolio →</CtaLink>
			<CtaLink href="/about" variant="secondary" tone="on-dark">About the archive</CtaLink>
		</div>
	</div>
</section>
