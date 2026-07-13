<script lang="ts">
	type Props = {
		items?: string[];
	};

	let { items = [] }: Props = $props();

	const row = $derived(items.length > 0 ? [...items, ...items] : []);
</script>

{#if items.length > 0}
	<div class="overflow-hidden border-b border-border/60 bg-ink text-cream">
		<div class="flex animate-marquee gap-12 whitespace-nowrap py-3 font-mono text-[11px] uppercase tracking-[0.3em]">
			{#each row as t}
				<span class="flex items-center gap-12 text-cream/80">
					{t}
					<span class="text-accent">✺</span>
				</span>
			{/each}
		</div>
	</div>
{/if}
