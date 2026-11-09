<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import { recordPageView } from '$lib/application/analytics';
	import ArtistAnswers from '$lib/components/artist/ArtistAnswers.svelte';
	import BlurhashImage from '$lib/components/BlurhashImage.svelte';
	import {
		acquisitionBlurhash,
		acquisitionImage,
		acquisitionTitle,
		artistDiscipline,
		artistHandle,
		artistName,
		artistPortrait,
		artistPortraitBlurhash
	} from '$lib/utils/fields';
	import { HEADER_ANSWER_KEYS, summarizeArtist } from '$lib/utils/artistSummary';

	let { data }: { data: PageData } = $props();

	const artist = $derived(data.artist);
	const galleryItems = $derived(data.posts);
	const others = $derived(data.others);
	const handle = $derived(artistHandle(artist));
	const firstWork = $derived(galleryItems[0]);
	const heroBlurhash = $derived(
		firstWork && acquisitionImage(firstWork) ? acquisitionBlurhash(firstWork) : artistPortraitBlurhash(artist)
	);
	const heroImage = $derived(
		firstWork ? acquisitionImage(firstWork) ?? artistPortrait(artist) : artistPortrait(artist)
	);

	const summary = $derived(summarizeArtist(artist, data.answers));
	const bodyAnswers = $derived(data.answers.filter((a) => !HEADER_ANSWER_KEYS.test(a.key)));

	onMount(() => recordPageView('artist', artist.id));
</script>

<svelte:head>
	<title>{artistName(artist)} — Artiv</title>
	<meta name="description" content={summary.tagline.slice(0, 160)} />
</svelte:head>

<section class="relative overflow-hidden border-b border-border/60 bg-ink text-cream">
	<div class="mx-auto grid max-w-[1600px] grid-cols-12 gap-6 px-6 py-12 md:px-10 md:py-20">
		<div class="col-span-12 md:col-span-7">
			<p class="flex items-center gap-3 font-mono text-[10px] uppercase tracking-[0.3em] text-accent">
				<span>●</span> Artist file / {artist.slug}
			</p>
			<h1 class="mt-6 font-display text-[13vw] leading-[0.92] tracking-tight md:text-[8vw]">
				{artistName(artist).split(' ')[0]}
				<br />
				<em class="not-italic text-cream/90">{artistName(artist).split(' ').slice(1).join(' ')}</em>
			</h1>
			<p class="mt-8 max-w-xl text-balance text-lg leading-relaxed text-cream/85">{summary.tagline}</p>
			<div class="mt-6">
				<a
					href="/@{handle}"
					class="inline-flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70 underline decoration-accent/50 underline-offset-4 transition hover:text-accent hover:decoration-accent"
				>
					Shareable profile · @{handle} →
				</a>
			</div>
			{#if summary.facts.length > 0}
				<div
					class="mt-10 grid max-w-lg grid-cols-2 gap-x-6 gap-y-4 border-t border-cream/15 pt-6 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70"
				>
					{#each summary.facts as f}
						<div><span class="text-cream/40">{f.label} · </span>{f.value}</div>
					{/each}
				</div>
			{/if}
		</div>
		<div class="col-span-12 md:col-span-5">
			<div class="grain relative aspect-[4/5] overflow-hidden rounded-sm bg-card">
				{#if heroImage}
					<BlurhashImage src={heroImage} alt={artistName(artist)} blurhash={heroBlurhash} loading="eager" class="h-full w-full object-cover" />
				{/if}
				{#if artistPortrait(artist)}
					<div class="absolute left-4 top-4 h-16 w-16 overflow-hidden rounded-full ring-2 ring-cream/70">
						<BlurhashImage src={artistPortrait(artist) ?? ""} alt="" blurhash={artistPortraitBlurhash(artist)} width={64} class="h-full w-full object-cover" />
					</div>
				{/if}
			</div>
			{#if summary.tags.length > 0}
				<div class="mt-4 flex flex-wrap gap-2">
					{#each summary.tags as i}
						<span
							class="rounded-full border border-cream/25 px-3 py-1 font-mono text-[10px] uppercase tracking-[0.15em] text-cream/80"
						>
							{i}
						</span>
					{/each}
				</div>
			{/if}
		</div>
	</div>
</section>

<ArtistAnswers answers={bodyAnswers} artistName={artistName(artist)} />

<section class="border-b border-border/60" aria-label="Works">
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-24">
		<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">Works</p>
		{#if galleryItems.length > 0}
			<ul class="mt-8 grid grid-cols-2 gap-4 md:grid-cols-3 md:gap-6">
				{#each galleryItems as item (item.id)}
					{@const image = acquisitionImage(item)}
					<li>
						<div class="grain aspect-[4/5] overflow-hidden rounded-sm bg-card">
							{#if image}
								<BlurhashImage
									src={image}
									alt={acquisitionTitle(item)}
									blurhash={acquisitionBlurhash(item)}
									loading="lazy"
									class="h-full w-full object-cover"
								/>
							{/if}
						</div>
						<p class="mt-3 font-display text-base italic text-foreground">{acquisitionTitle(item)}</p>
						{#if item.medium || item.year}
							<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
								{[item.medium, item.year].filter(Boolean).join(' · ')}
							</p>
						{/if}
					</li>
				{/each}
			</ul>
		{:else}
			<p class="mt-4 font-display text-2xl text-foreground md:text-3xl">No works published yet.</p>
		{/if}
	</div>
</section>

<section class="bg-card/40">
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-24">
		<div class="flex items-baseline justify-between gap-6">
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
				Continue the conversation
			</p>
			<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				End of file · {artistName(artist)}
			</p>
		</div>
		{#if others.length > 0}
			<div class="mt-10 grid gap-6 sm:grid-cols-3">
				{#each others as a}
					<a
						href="/artists/{a.slug}"
						class="group flex items-center gap-4 border-t border-border/60 pt-5"
					>
						<div class="h-16 w-16 shrink-0 overflow-hidden rounded-sm bg-muted">
							{#if artistPortrait(a)}
								<BlurhashImage
									src={artistPortrait(a) ?? ''}
									alt={artistName(a)}
									blurhash={artistPortraitBlurhash(a)}
									width={64}
									loading="lazy"
									class="h-full w-full object-cover grayscale transition group-hover:grayscale-0"
								/>
							{/if}
						</div>
						<div>
							<p class="font-display text-lg text-foreground">{artistName(a)}</p>
							<p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
								{artistDiscipline(a) ?? ''}
							</p>
						</div>
						<span
							class="ml-auto font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground transition group-hover:text-accent"
						>
							→
						</span>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</section>
