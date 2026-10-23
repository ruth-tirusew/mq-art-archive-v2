<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '../cn';

	type Props = {
		eyebrow?: string;
		title: string;
		body?: string;
		/** The next step. Every empty state should have one when the user can act. */
		action?: Snippet;
		class?: string;
	};

	let { eyebrow = 'Nothing here yet', title, body, action, class: className }: Props = $props();
</script>

<div class={cn('rounded-sm border border-dashed border-border/70 bg-card/30 p-8 md:p-12', className)}>
	<p class="t-eyebrow text-accent">{eyebrow}</p>
	<h3 class="t-h3 mt-4 max-w-lg text-foreground md:text-3xl">{title}</h3>
	{#if body}<p class="t-body mt-4 max-w-xl text-muted-foreground">{body}</p>{/if}
	{#if action}<div class="mt-6 flex flex-wrap gap-3">{@render action()}</div>{/if}
</div>
