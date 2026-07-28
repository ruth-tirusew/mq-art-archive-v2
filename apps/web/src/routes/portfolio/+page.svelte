<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
</script>

<svelte:head>
	<title>Portfolio Builder — Artiv</title>
	<meta
		name="description"
		content="A link-in-bio built for Ethiopian creatives — portfolio, contact, and Telegram in one shareable link."
	/>
</svelte:head>

<section class="border-b border-border/60">
	<div class="mx-auto grid max-w-[1600px] gap-12 px-6 py-14 md:grid-cols-12 md:px-10 md:py-20">
		<div class="md:col-span-7">
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
				⁂ &nbsp; Pillar 02 · Digital identity
			</p>
			<h1 class="mt-4 max-w-xl font-display text-4xl leading-[1.05] text-foreground md:text-6xl">
				One link. The whole <em class="italic">studio</em>.
			</h1>
			<p class="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg">
				A focused, shareable profile built for painters, musicians, writers and artisans —
				surfacing Telegram, Instagram, and direct contact links so a gallerist or client can reach
				you in two taps.
			</p>

			<div class="mt-8 flex flex-wrap items-center gap-3">
				<CtaLink href="/apply" variant="primary">Claim your @handle →</CtaLink>
				{#if data.handle}
					<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
						artiv.et / @{data.handle}
					</span>
				{/if}
			</div>

			{#if data.handle}
				<a
					href="/@{data.handle}"
					class="mt-4 inline-block font-mono text-[10px] uppercase tracking-[0.2em] text-accent underline decoration-accent underline-offset-4 hover:text-foreground"
				>
					Open featured artist →
				</a>
			{/if}

			<ul class="mt-10 grid grid-cols-2 gap-x-6 gap-y-3 text-sm text-foreground/90">
				{#each [
					'Verified handle + tier badge',
					'Six-work portfolio strip',
					'Telegram, IG & WhatsApp',
					'Direct contact links',
					'Press kit + high-res download',
					'QR for in-person handoff'
				] as f}
					<li class="flex items-start gap-2.5">
						<span class="mt-1.5 inline-block h-1.5 w-1.5 shrink-0 rounded-full bg-accent"></span>
						{f}
					</li>
				{/each}
			</ul>
		</div>

		<div class="md:col-span-5">
			{#if data.artist}
				<div class="transition hover:opacity-95">
					<ShareableProfile artist={data.artist} works={data.posts} demo framed />
				</div>
			{:else}
				<div class="flex min-h-[420px] items-center justify-center rounded-sm border border-border bg-card/40 p-8 text-center">
					<p class="max-w-xs text-sm leading-relaxed text-muted-foreground">
						The featured artist preview is temporarily unavailable.
					</p>
				</div>
			{/if}
		</div>
	</div>
</section>
