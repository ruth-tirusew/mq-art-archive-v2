<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import {
		acquisitionArtistName,
		acquisitionArtistSlug,
		acquisitionImage,
		acquisitionMedium,
		acquisitionTitle
	} from '$lib/utils/fields';
	import type { ArtistTile, CuratorModule, EditorialSpread } from './editorialCompositions';

	type Props = {
		spreads: EditorialSpread[];
	};

	let { spreads }: Props = $props();

	let index = $state(0);
	const spread = $derived(spreads[index] ?? spreads[0]);
	const maxIndex = $derived(Math.max(0, spreads.length - 1));

	function prev() {
		index = index <= 0 ? maxIndex : index - 1;
	}

	function next() {
		index = index >= maxIndex ? 0 : index + 1;
	}

	function artistHref(post: ArtPost) {
		const slug = acquisitionArtistSlug(post);
		return slug ? `/artists/${slug}` : '/archive';
	}

	function label(post: ArtPost) {
		return acquisitionArtistName(post) || acquisitionTitle(post);
	}

	function medium(post: ArtPost) {
		return acquisitionMedium(post) ?? '';
	}
</script>

{#if spread}
	<section class="curation relative mx-auto w-full max-w-[1600px] px-5 py-5 md:px-10 md:py-6">
		{#key index}
			<div class="spread" data-layout={spread.layoutId}>
				<!-- Intro claim -->
				<div class="cell cell-intro flex flex-col justify-center py-1 md:py-4 md:pr-4 lg:pr-8">
					<p class="font-mono text-[10px] uppercase tracking-[0.28em] text-accent">
						{spread.intro.eyebrow}
					</p>
					<h1
						class="mt-3 max-w-[14ch] font-display text-[1.75rem] leading-[1.12] tracking-tight text-foreground sm:max-w-[16ch] md:mt-4 md:text-[2.15rem] lg:text-[2.55rem]"
					>
						{spread.intro.headline}
					</h1>
					<p
						class="intro-body mt-4 max-w-[36ch] text-sm leading-relaxed text-muted-foreground md:text-[15px]"
					>
						{spread.intro.body}
					</p>
					<a
						href={spread.intro.ctaHref}
						class="mt-5 inline-flex w-fit font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 transition hover:text-accent md:mt-6"
					>
						{spread.intro.ctaLabel} →
					</a>
				</div>

				{@render artistTile(spread.feature, 'cell-feature')}
				{@render artistTile(spread.archive, 'cell-archive')}
				{@render artistTile(spread.support, 'cell-support')}
				{@render noteTile(spread.note)}
				{@render artistTile(spread.emerging, 'cell-emerging')}
			</div>
		{/key}

		<div class="spread-footer mt-5 flex flex-wrap items-center justify-between gap-4 md:mt-6">
			<div class="spread-pager flex items-center gap-3">
				<button
					type="button"
					onclick={prev}
					disabled={spreads.length <= 1}
					aria-label="Previous curation"
					class="flex h-9 w-9 items-center justify-center border border-foreground/20 text-foreground transition hover:border-foreground hover:bg-foreground hover:text-cream disabled:pointer-events-none disabled:opacity-30"
				>
					←
				</button>
				<div class="flex items-center gap-2" role="tablist" aria-label="Curations">
					{#each spreads as _, i}
						<button
							type="button"
							role="tab"
							aria-selected={i === index}
							aria-label="Curation {i + 1}"
							onclick={() => (index = i)}
							class="h-2 w-2 rounded-full transition {i === index
								? 'scale-110 bg-accent'
								: 'bg-foreground/25 hover:bg-foreground/50'}"
						></button>
					{/each}
				</div>
				<button
					type="button"
					onclick={next}
					disabled={spreads.length <= 1}
					aria-label="Next curation"
					class="flex h-9 w-9 items-center justify-center border border-foreground/20 text-foreground transition hover:border-foreground hover:bg-foreground hover:text-cream disabled:pointer-events-none disabled:opacity-30"
				>
					→
				</button>
			</div>

			<a
				href="/archive"
				class="browse-archive font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 transition hover:text-accent"
			>
				Browse the archive →
			</a>
		</div>
	</section>
{/if}

{#snippet artistTile(tile: ArtistTile, cellClass: string)}
	{@const post = tile.post}
	<a href={artistHref(post)} class="group relative min-h-0 overflow-hidden bg-muted {cellClass}">
		<img
			src={acquisitionImage(post) ?? ''}
			alt={acquisitionTitle(post)}
			class="h-full w-full object-cover transition duration-700 ease-out group-hover:scale-[1.03]"
			loading={cellClass === 'cell-feature' ? 'eager' : 'lazy'}
		/>
		<div
			class="pointer-events-none absolute inset-0 bg-gradient-to-t from-ink/70 via-ink/10 to-transparent"
		></div>

		{#if tile.badge}
			<span
				class="absolute left-3 top-3 bg-cream px-2.5 py-1 font-mono text-[9px] uppercase tracking-[0.22em] text-foreground md:left-4 md:top-4"
			>
				{tile.badge}
			</span>
		{/if}

		<div
			class="absolute inset-x-0 bottom-0 flex items-end justify-between gap-2 p-3 md:gap-3 md:p-4 lg:p-5"
		>
			<div class="min-w-0">
				<p
					class="truncate font-display text-[0.95rem] leading-tight text-cream md:text-lg lg:text-xl"
				>
					{label(post)}
				</p>
				{#if medium(post)}
					<p
						class="mt-1 truncate font-mono text-[9px] uppercase tracking-[0.16em] text-cream/75 md:text-[10px] md:tracking-[0.18em]"
					>
						{medium(post)}
					</p>
				{/if}
			</div>
			<span
				class="tile-cta shrink-0 font-mono text-[10px] uppercase tracking-[0.16em] text-cream/90 transition group-hover:text-cream"
			>
				View artist →
			</span>
		</div>
	</a>
{/snippet}

{#snippet noteTile(note: CuratorModule)}
	<div class="cell-note flex flex-col justify-center py-5 md:px-2 md:py-4 lg:px-4">
		<p class="font-mono text-[10px] uppercase tracking-[0.28em] text-accent">{note.eyebrow}</p>
		<h2
			class="mt-3 max-w-[16ch] font-display text-[1.55rem] leading-[1.15] tracking-tight text-foreground md:mt-4 md:max-w-[18ch] md:text-[1.65rem] lg:text-[1.85rem]"
		>
			{note.title}
		</h2>
		<p class="note-body mt-3 max-w-[34ch] text-sm leading-relaxed text-muted-foreground">
			{note.body}
		</p>
		<a
			href={note.href}
			class="mt-4 inline-flex w-fit font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent decoration-2 underline-offset-8 transition hover:text-accent md:mt-5"
		>
			{note.cta} →
		</a>
	</div>
{/snippet}

<style>
	.spread {
		display: grid;
		gap: 0.65rem;
		grid-template-columns: 1fr 1fr;
		animation: spread-in 520ms cubic-bezier(0.22, 1, 0.36, 1) both;
	}

	@keyframes spread-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	/* ——— Mobile-first stack (matches mockup) ——— */
	.spread .cell-intro {
		grid-column: 1 / -1;
	}

	.spread .intro-body,
	.spread .note-body {
		display: none;
	}

	.spread .cell-feature {
		grid-column: 1 / -1;
		aspect-ratio: 4 / 5;
		min-height: 0;
	}

	.spread .cell-archive {
		grid-column: 1;
		aspect-ratio: 1;
		min-height: 0;
	}

	.spread .cell-support {
		grid-column: 2;
		aspect-ratio: 1;
		min-height: 0;
	}

	.spread .cell-archive .tile-cta,
	.spread .cell-support .tile-cta,
	.spread .cell-emerging .tile-cta {
		display: none;
	}

	.spread .cell-note {
		grid-column: 1 / -1;
		padding-top: 0.75rem;
		padding-bottom: 0.75rem;
	}

	.spread .cell-emerging {
		grid-column: 1 / -1;
		aspect-ratio: 16 / 10;
		min-height: 0;
	}

	.spread-footer {
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		margin-top: 1.25rem;
	}

	.spread-pager {
		order: 2;
	}

	.browse-archive {
		order: 1;
		text-align: center;
	}

	/* ——— Desktop compositions ——— */
	@media (min-width: 768px) {
		.spread {
			gap: 0.75rem;
		}

		.spread .intro-body,
		.spread .note-body {
			display: block;
		}

		.spread .cell-archive .tile-cta,
		.spread .cell-support .tile-cta,
		.spread .cell-emerging .tile-cta {
			display: inline;
		}

		.spread .cell-feature,
		.spread .cell-archive,
		.spread .cell-support,
		.spread .cell-emerging {
			aspect-ratio: auto;
			min-height: 0;
		}

		.spread-footer {
			flex-direction: row;
			align-items: center;
			justify-content: space-between;
			margin-top: 1.5rem;
		}

		.spread-pager,
		.browse-archive {
			order: unset;
			text-align: left;
		}

		/* Editorial A — mockup */
		.spread[data-layout='editorial-a'] {
			grid-template-columns: repeat(12, minmax(0, 1fr));
			grid-template-rows: minmax(300px, 54vh) minmax(240px, 36vh);
		}
		.spread[data-layout='editorial-a'] .cell-intro {
			grid-column: 1 / 4;
			grid-row: 1;
		}
		.spread[data-layout='editorial-a'] .cell-feature {
			grid-column: 4 / 11;
			grid-row: 1;
		}
		.spread[data-layout='editorial-a'] .cell-support {
			grid-column: 11 / 13;
			grid-row: 1;
		}
		.spread[data-layout='editorial-a'] .cell-archive {
			grid-column: 1 / 4;
			grid-row: 2;
		}
		.spread[data-layout='editorial-a'] .cell-note {
			grid-column: 4 / 7;
			grid-row: 2;
			padding-top: 0;
			padding-bottom: 0;
		}
		.spread[data-layout='editorial-a'] .cell-emerging {
			grid-column: 7 / 13;
			grid-row: 2;
		}

		/* Editorial B */
		.spread[data-layout='editorial-b'] {
			grid-template-columns: repeat(12, minmax(0, 1fr));
			grid-template-rows: auto minmax(340px, 48vh) minmax(220px, 34vh);
		}
		.spread[data-layout='editorial-b'] .cell-intro {
			grid-column: 1 / 13;
			grid-row: 1;
			max-width: 40rem;
			padding-bottom: 0.5rem;
		}
		.spread[data-layout='editorial-b'] .cell-feature {
			grid-column: 1 / 13;
			grid-row: 2;
		}
		.spread[data-layout='editorial-b'] .cell-archive {
			grid-column: 1 / 4;
			grid-row: 3;
		}
		.spread[data-layout='editorial-b'] .cell-note {
			grid-column: 4 / 8;
			grid-row: 3;
			padding-top: 0;
			padding-bottom: 0;
		}
		.spread[data-layout='editorial-b'] .cell-emerging {
			grid-column: 8 / 13;
			grid-row: 3;
		}
		.spread[data-layout='editorial-b'] .cell-support {
			display: none;
		}

		/* Editorial C */
		.spread[data-layout='editorial-c'] {
			grid-template-columns: repeat(12, minmax(0, 1fr));
			grid-template-rows: auto minmax(320px, 52vh) minmax(200px, 32vh);
		}
		.spread[data-layout='editorial-c'] .cell-intro {
			grid-column: 1 / 13;
			grid-row: 1;
			max-width: 36rem;
			padding-bottom: 0.25rem;
		}
		.spread[data-layout='editorial-c'] .cell-feature {
			grid-column: 1 / 9;
			grid-row: 2;
		}
		.spread[data-layout='editorial-c'] .cell-support {
			grid-column: 9 / 13;
			grid-row: 2;
			display: block;
		}
		.spread[data-layout='editorial-c'] .cell-archive {
			grid-column: 1 / 5;
			grid-row: 3;
		}
		.spread[data-layout='editorial-c'] .cell-note {
			grid-column: 5 / 9;
			grid-row: 3;
			padding-top: 0;
			padding-bottom: 0;
		}
		.spread[data-layout='editorial-c'] .cell-emerging {
			grid-column: 9 / 13;
			grid-row: 3;
		}
	}
</style>
