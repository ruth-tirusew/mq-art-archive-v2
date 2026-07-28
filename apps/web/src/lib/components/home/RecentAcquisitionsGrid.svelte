<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import AcquisitionCard from './AcquisitionCard.svelte';

	type Props = {
		items: ArtPost[];
	};

	let { items }: Props = $props();

	const pageSize = 6;
	let page = $state(0);
	const maxPage = $derived(Math.max(0, Math.ceil(items.length / pageSize) - 1));

	const pageWorks = $derived(items.slice(page * pageSize, page * pageSize + pageSize));
	const featured = $derived(pageWorks[0]);
	const rest = $derived(pageWorks.slice(1));
	const stacked = $derived(rest.slice(0, 2));
	const row = $derived(rest.slice(2, 5));
</script>

{#if featured}
	<div class="relative mt-10 max-h-[100vh]">
		<button
			type="button"
			onclick={() => (page = Math.max(0, page - 1))}
			disabled={page === 0}
			aria-label="Previous acquisitions"
			class="absolute -left-2 top-1/2 z-10 hidden h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-foreground/85 text-lg text-cream shadow-lg transition hover:bg-foreground disabled:pointer-events-none disabled:opacity-30 md:flex lg:-left-14"
		>
			‹
		</button>

		<button
			type="button"
			onclick={() => (page = Math.min(maxPage, page + 1))}
			disabled={page >= maxPage}
			aria-label="Next acquisitions"
			class="absolute -right-2 top-1/2 z-10 hidden h-11 w-11 -translate-y-1/2 items-center justify-center rounded-full bg-foreground/85 text-lg text-cream shadow-lg transition hover:bg-foreground disabled:pointer-events-none disabled:opacity-30 md:flex lg:-right-14"
		>
			›
		</button>

		<div class="flex max-h-[100vh] min-h-0 flex-col gap-5 overflow-hidden md:gap-6">
			<div class="grid min-h-0 flex-[3] grid-cols-1 items-stretch gap-5 md:grid-cols-12 md:gap-6">
				<div class="min-h-0 md:col-span-7">
					<AcquisitionCard item={featured} featured class="h-full" />
				</div>
				<div class="flex min-h-0 flex-col gap-5 md:col-span-5 md:gap-6">
					{#each stacked as w}
						<AcquisitionCard item={w} compact class="min-h-0 flex-1" />
					{/each}
				</div>
			</div>

			{#if row.length > 0}
				<div class="grid min-h-0 flex-[2] grid-cols-1 gap-5 sm:grid-cols-3 md:gap-6">
					{#each row as w}
						<AcquisitionCard item={w} class="min-h-0" />
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/if}
