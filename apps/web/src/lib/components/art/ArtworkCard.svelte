<script lang="ts">
	import BlurhashImage from '$lib/components/BlurhashImage.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import { cn } from '$lib/utils';
	import { htmlText, truncateText } from '$lib/utils/answers';
	import {
		acquisitionArtistName,
		acquisitionBlurhash,
		acquisitionDimensions,
		acquisitionImage,
		acquisitionMedium,
		acquisitionPalette,
		acquisitionTitle,
		acquisitionYear
	} from '$lib/utils/fields';

	type Props = {
		item: ArtPost;
		/** Where the card goes. Without one it is not a link (hover details still show). */
		href?: string;
		/** Open `href` in a new tab, e.g. a full-size image. */
		external?: boolean;
		/** Name the artist on the card. Leave off where the artist is already obvious. */
		showArtist?: boolean;
		/**
		 * The artist's own page of work: the hover panel also lists year, medium, size and place,
		 * shows the palette as swatches, and gives a short excerpt of the description.
		 */
		detailed?: boolean;
		loading?: 'lazy' | 'eager';
		/** Called if the image fails to load, so a list can drop a dead tile. */
		onerror?: () => void;
		class?: string;
	};

	let {
		item,
		href,
		external = false,
		showArtist = false,
		detailed = false,
		loading = 'lazy',
		onerror,
		class: className
	}: Props = $props();

	const image = $derived(acquisitionImage(item));
	const title = $derived(acquisitionTitle(item));
	const artist = $derived(showArtist ? acquisitionArtistName(item) : '');
	const palette = $derived(acquisitionPalette(item));
	const year = $derived(acquisitionYear(item));
	const medium = $derived(acquisitionMedium(item));
	const summary = $derived([medium, year].filter(Boolean).join(' · '));
	const caption = $derived((showArtist ? [artist, year] : [medium, year]).filter(Boolean).join(' · '));

	// Only what the artist actually gave: an empty row would just be noise.
	const details = $derived(
		detailed
			? [
					{ label: 'Year', value: year ? String(year) : '' },
					{ label: 'Medium', value: medium ?? '' },
					{ label: 'Size', value: acquisitionDimensions(item) ?? '' },
					{ label: 'Place', value: item.city ?? '' }
				].filter((d) => d.value)
			: []
	);
	const blurb = $derived(detailed && item.description ? truncateText(htmlText(item.description), 120) : '');
</script>

{#snippet card()}
	<div class="grain relative aspect-[4/5] overflow-hidden rounded-sm bg-card">
		{#if image}
			<BlurhashImage
				src={image}
				alt={title}
				blurhash={acquisitionBlurhash(item)}
				{loading}
				{onerror}
				class="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.06] group-focus-visible:scale-[1.06]"
			/>
		{/if}
		<div class="absolute inset-0 bg-ink/0 transition duration-500 group-hover:bg-ink/25 group-focus-visible:bg-ink/25"></div>

		<!-- Details on hover or keyboard focus. The text is always in the page, so it is also read aloud. -->
		<div
			class="absolute inset-x-0 bottom-0 translate-y-3 p-4 opacity-0 transition duration-500 group-hover:translate-y-0 group-hover:opacity-100 group-focus-visible:translate-y-0 group-focus-visible:opacity-100"
		>
			<div class="bg-gradient-to-t from-ink/95 via-ink/80 to-transparent p-4">
				<p class="font-display text-lg italic text-cream">{title}</p>
				{#if artist}
					<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/80">{artist}</p>
				{/if}
				{#if detailed}
					{#if details.length > 0}
						<dl class="mt-2 grid grid-cols-[auto_1fr] gap-x-3 gap-y-0.5 font-mono text-[10px] uppercase tracking-[0.18em]">
							{#each details as d (d.label)}
								<dt class="text-cream/50">{d.label}</dt>
								<dd class="text-cream/85">{d.value}</dd>
							{/each}
						</dl>
					{/if}
					{#if blurb}
						<p class="mt-2 line-clamp-2 text-xs leading-relaxed text-cream/75">{blurb}</p>
					{/if}
					{#if palette.length > 0}
						<ul class="mt-2.5 flex gap-1.5" aria-label="Palette">
							{#each palette as c (c)}
								<li class="h-3.5 w-3.5 rounded-full ring-1 ring-cream/40" style:background={c} title={c}></li>
							{/each}
						</ul>
					{/if}
				{:else if summary}
					<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/60">{summary}</p>
				{/if}
				{#if href}
					<span class="mt-2 inline-block font-mono text-[10px] uppercase tracking-[0.2em] text-accent">
						{external ? 'View larger' : 'View'} →
					</span>
				{/if}
			</div>
		</div>
	</div>

	<!-- Always visible, so touch screens (no hover) still get the basics. Fades out while the panel is up. -->
	<div class="mt-4 flex items-start justify-between gap-4 transition group-hover:opacity-0 group-focus-visible:opacity-0">
		<div class="min-w-0">
			<p class="truncate font-display text-lg italic text-foreground">{title}</p>
			{#if caption}
				<p class="truncate font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{caption}</p>
			{/if}
		</div>
		{#if palette.length > 0}
			<div class="flex shrink-0 gap-1 pt-1" aria-hidden="true">
				{#each palette.slice(0, 5) as c (c)}
					<span class="h-5 w-1.5" style:background={c}></span>
				{/each}
			</div>
		{/if}
	</div>
{/snippet}

{#if href}
	<a
		{href}
		target={external ? '_blank' : undefined}
		rel={external ? 'noopener noreferrer' : undefined}
		class={cn('group block focus-visible:outline-none', className)}
	>
		{@render card()}
	</a>
{:else}
	<div class={cn('group block', className)}>
		{@render card()}
	</div>
{/if}
