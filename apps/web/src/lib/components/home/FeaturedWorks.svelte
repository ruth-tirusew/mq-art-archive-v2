<script lang="ts">
	import ArtworkCard from '$lib/components/art/ArtworkCard.svelte';
	import SectionHeader from '$lib/components/home/SectionHeader.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import { acquisitionArtistSlug, acquisitionImage } from '$lib/utils/fields';

	type Props = {
		works: ArtPost[];
		/** False when these are simply the latest works because nothing has been featured. */
		featured?: boolean;
		/** Works by this artist are already shown in the featured-artist section. */
		excludeArtistSlug?: string;
	};

	let { works, excludeArtistSlug, featured = true }: Props = $props();

	// Seeded/legacy media can point at dead URLs; drop those tiles rather than show a broken image.
	let failed = $state<string[]>([]);

	const items = $derived(
		works
			.filter((w) => !excludeArtistSlug || acquisitionArtistSlug(w) !== excludeArtistSlug)
			.map((w) => ({ work: w, src: acquisitionImage(w) }))
			.filter((w): w is { work: ArtPost; src: string } => Boolean(w.src) && !failed.includes(w.work.id))
			.slice(0, 8)
	);
</script>

{#if items.length > 0}
	<section class="bg-background" aria-labelledby="works-heading">
		<div class="mx-auto flex max-w-[1600px] flex-col gap-12 px-6 py-20 md:px-10 md:py-28">
			<SectionHeader eyebrow="From the archive" id="works-heading" action={{ href: '/archive', label: 'Browse the archive' }}>
				{featured ? 'Featured' : 'Recent'} <em class="italic">works.</em>
			</SectionHeader>

			<div class="-mx-6 flex snap-x scroll-px-6 gap-6 overflow-x-auto px-6 pb-2 md:-mx-10 md:scroll-px-10 md:px-10 [scrollbar-width:none]">
				{#each items as { work } (work.id)}
					<ArtworkCard
						item={work}
						href={acquisitionArtistSlug(work) ? `/artists/${acquisitionArtistSlug(work)}` : '/archive'}
						showArtist
						onerror={() => (failed = [...failed, work.id])}
						class="w-[72vw] shrink-0 snap-start sm:w-[44vw] lg:w-[calc((100%-4.5rem)/4)]"
					/>
				{/each}
			</div>
		</div>
	</section>
{/if}
