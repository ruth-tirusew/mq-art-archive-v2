<script lang="ts">
	type Props = {
		items?: string[];
	};

	let { items = [] }: Props = $props();

	const row = $derived(items.length > 0 ? [...items, ...items] : []);
</script>

{#if items.length > 0}
	<div class="flex items-center gap-6 border-y border-border/60 bg-background py-5 md:gap-10">
		<p class="t-eyebrow shrink-0 pl-6 text-accent md:pl-10">Coming up</p>
		<div class="marquee-mask min-w-0 flex-1 overflow-hidden">
			<div class="flex w-max animate-marquee gap-14 font-display text-xl whitespace-nowrap text-foreground md:text-2xl">
				{#each row as t}
					<span class="flex items-center gap-14">
						{t}
						<span class="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden="true"></span>
					</span>
				{/each}
			</div>
		</div>
	</div>
{/if}

<style>
	.marquee-mask {
		mask-image: linear-gradient(to right, transparent, black 8%, black 92%, transparent);
	}
</style>
