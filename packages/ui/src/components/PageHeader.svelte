<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '../cn';

	type Props = {
		eyebrow: string;
		/** Wrap one word in <em class="italic"> for the editorial emphasis. */
		children: Snippet;
		lead?: string;
		actions?: Snippet;
		class?: string;
	};

	let { eyebrow, children, lead, actions, class: className }: Props = $props();
</script>

<header class={cn('flex flex-wrap items-end justify-between gap-x-10 gap-y-6', className)}>
	<div class="flex max-w-3xl flex-col gap-4">
		<p class="t-eyebrow text-accent">{eyebrow}</p>
		<h1 class="t-h1 text-foreground">{@render children()}</h1>
		{#if lead}<p class="t-lead text-muted-foreground">{lead}</p>{/if}
	</div>
	{#if actions}<div class="flex flex-wrap items-center gap-3">{@render actions()}</div>{/if}
</header>
