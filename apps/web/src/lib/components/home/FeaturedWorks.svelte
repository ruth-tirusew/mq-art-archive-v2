<script lang="ts">
	import SectionHeader from '$lib/components/home/SectionHeader.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import {
		acquisitionArtistName,
		acquisitionArtistSlug,
		acquisitionImage,
		acquisitionMedium,
		acquisitionYear
	} from '$lib/utils/fields';

	type Props = {
		works: ArtPost[];
		/** Works by this artist are already shown in the featured-artist section. */
		excludeArtistSlug?: string;
	};

	let { works, excludeArtistSlug }: Props = $props();

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
				Featured <em class="italic">works.</em>
			</SectionHeader>

			<div class="-mx-6 flex snap-x scroll-px-6 gap-6 overflow-x-auto px-6 pb-2 md:-mx-10 md:scroll-px-10 md:px-10 [scrollbar-width:none]">
				{#each items as { work, src } (work.id)}
					<a
						href={acquisitionArtistSlug(work) ? `/artists/${acquisitionArtistSlug(work)}` : '/archive'}
						class="group flex w-[72vw] shrink-0 snap-start flex-col gap-4 sm:w-[44vw] lg:w-[calc((100%-4.5rem)/4)]"
					>
						<div class="aspect-[4/5] overflow-hidden rounded-sm bg-muted">
							<img
								{src}
								alt={work.title}
								loading="lazy"
								onerror={() => (failed = [...failed, work.id])}
								class="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.03]"
							/>
						</div>
						<div class="flex flex-col gap-1">
							<p class="t-h3 truncate text-foreground group-hover:text-accent">{work.title}</p>
							<p class="text-sm text-muted-foreground">{acquisitionArtistName(work, 'Unknown artist')}</p>
							<p class="t-eyebrow text-muted-foreground">
								{[acquisitionMedium(work), acquisitionYear(work)].filter(Boolean).join(' · ')}
							</p>
						</div>
					</a>
				{/each}
			</div>
		</div>
	</section>
{/if}
