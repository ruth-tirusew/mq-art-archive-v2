<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';
	import { recordPageView } from '$lib/application/analytics';
	import ArtistTimeline from '$lib/components/artist/ArtistTimeline.svelte';
	import {
		acquisitionImage,
		acquisitionPalette,
		acquisitionTitle,
		artistBorn,
		artistDiscipline,
		artistHandle,
		artistInfluences,
		artistLocation,
		artistName,
		artistPortrait,
		artistTagline,
		artistYearsActive
	} from '$lib/utils/fields';

	let { data }: { data: PageData } = $props();

	const artist = $derived(data.artist);
	const galleryItems = $derived(data.posts);
	const others = $derived(data.others);
	const handle = $derived(artistHandle(artist));
	const firstWork = $derived(galleryItems[0]);
	const heroImage = $derived(
		firstWork ? acquisitionImage(firstWork) ?? artistPortrait(artist) : artistPortrait(artist)
	);

	const timeline = $derived(
		galleryItems.map((item) => {
			const image = acquisitionImage(item);
			const title = acquisitionTitle(item);
			return {
				year: item.year ?? new Date().getFullYear(),
				kind: 'work' as const,
				title,
				detail: item.medium,
				work: image
					? {
							image,
							title,
							palette: acquisitionPalette(item)
						}
					: undefined
			};
		})
	);

	onMount(() => recordPageView('artist', artist.id));
</script>

<svelte:head>
	<title>{artistName(artist)} — Artiv</title>
	<meta name="description" content={(artistTagline(artist) || '').slice(0, 160)} />
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
			<p class="mt-8 max-w-xl text-balance text-lg leading-relaxed text-cream/85">{artistTagline(artist)}</p>
			<div class="mt-6">
				<a
					href="/@{handle}"
					class="inline-flex items-center gap-2 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70 underline decoration-accent/50 underline-offset-4 transition hover:text-accent hover:decoration-accent"
				>
					Shareable profile · @{handle} →
				</a>
			</div>
			<div
				class="mt-10 grid max-w-lg grid-cols-2 gap-x-6 gap-y-4 border-t border-cream/15 pt-6 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70"
			>
				<div><span class="text-cream/40">Born · </span>{artistBorn(artist) ?? '—'}</div>
				<div><span class="text-cream/40">Based · </span>{artistLocation(artist) || '—'}</div>
				<div><span class="text-cream/40">Discipline · </span>{artistDiscipline(artist) ?? '—'}</div>
				<div><span class="text-cream/40">Active · </span>{artistYearsActive(artist) ?? '—'}</div>
			</div>
		</div>
		<div class="col-span-12 md:col-span-5">
			<div class="grain relative aspect-[4/5] overflow-hidden rounded-sm bg-card">
				{#if heroImage}
					<img src={heroImage} alt={artistName(artist)} class="h-full w-full object-cover" />
				{/if}
				<div class="absolute left-4 top-4 h-16 w-16 overflow-hidden rounded-full ring-2 ring-cream/70">
					{#if artistPortrait(artist)}
						<img src={artistPortrait(artist)} alt="" class="h-full w-full object-cover" />
					{/if}
				</div>
			</div>
			{#if artistInfluences(artist).length > 0}
				<div class="mt-4 flex flex-wrap gap-2">
					{#each artistInfluences(artist) as i}
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

{#if timeline.length > 0}
	<ArtistTimeline entries={timeline} artistName={artistName(artist)} />
{:else}
	<section class="border-b border-border/60">
		<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-24">
			<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">Works</p>
			<p class="mt-4 font-display text-2xl text-foreground md:text-3xl">
				No works archived for this file yet.
			</p>
			<p class="mt-3 max-w-md text-sm leading-relaxed text-muted-foreground">
				Check back as new acquisitions land, or visit the shareable profile for contact.
			</p>
		</div>
	</section>
{/if}

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
								<img
									src={artistPortrait(a)}
									alt={artistName(a)}
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
