<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import {
		acquisitionArtistName,
		acquisitionArtistSlug,
		acquisitionDimensions,
		acquisitionFeatured,
		acquisitionImage,
		acquisitionMedium,
		acquisitionPalette,
		acquisitionTitle,
		acquisitionYear
	} from '$lib/utils/fields';

	type Props = {
		item: ArtPost;
		artistName?: string;
		featured?: boolean;
		compact?: boolean;
		class?: string;
	};

	let {
		item,
		artistName: artistNameProp = '',
		featured = false,
		compact = false,
		class: className = ''
	}: Props = $props();

	const label = $derived(acquisitionArtistName(item, artistNameProp) || artistNameProp);
</script>

<a
	href="/artists/{acquisitionArtistSlug(item)}"
	class="group flex min-h-0 flex-col overflow-hidden bg-card transition duration-300 hover:-translate-y-0.5 hover:shadow-[0_16px_40px_-18px_rgba(0,0,0,0.28)] {className}"
>
	<div
		class="grain overflow-hidden bg-[#ebe6dc] {featured
			? 'aspect-[16/10] shrink-0'
			: compact
				? 'relative min-h-0 flex-1'
				: 'aspect-[4/5] shrink-0 md:min-h-0 md:flex-1'}"
	>
		<img
			src={acquisitionImage(item) ?? ''}
			alt={acquisitionTitle(item)}
			loading={featured ? 'eager' : 'lazy'}
			class="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.03]"
		/>
	</div>

	<div
		class="grid min-h-[5.75rem] grid-cols-[minmax(0,1fr)_auto_auto] items-end gap-3 border-t border-border/30 bg-card px-4 py-4 md:gap-5 md:px-5"
	>
		<div class="min-w-0">
			{#if acquisitionFeatured(item)}
				<p class="mb-1.5 font-mono text-[10px] uppercase tracking-[0.28em] text-accent">
					New acquisition
				</p>
			{/if}
			<p
				class="truncate font-display font-medium leading-tight text-foreground {featured
					? 'text-xl md:text-2xl'
					: 'text-lg'}"
			>
				{featured ? label : acquisitionTitle(item)}
			</p>
			<p class="mt-1 truncate font-mono text-[10px] uppercase tracking-[0.18em] text-muted-foreground">
				{label}{#if acquisitionYear(item)} · {acquisitionYear(item)}{/if}
			</p>
		</div>

		<div class="space-y-1 text-right font-mono text-[10px] uppercase leading-snug tracking-[0.14em] text-muted-foreground">
			{#if acquisitionMedium(item)}
				<p class="whitespace-nowrap">{acquisitionMedium(item)}</p>
			{/if}
			{#if acquisitionDimensions(item)}
				<p class="whitespace-nowrap">{acquisitionDimensions(item)}</p>
			{/if}
		</div>

		<div class="flex gap-1 pb-0.5">
			{#each acquisitionPalette(item).slice(0, 3) as c}
				<span class="h-9 w-1.5 shrink-0" style:background={c}></span>
			{/each}
		</div>
	</div>
</a>
