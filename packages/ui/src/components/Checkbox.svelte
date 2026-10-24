<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '../cn';

	type Props = Omit<HTMLInputAttributes, 'type' | 'children'> & { class?: string; children: Snippet };

	let { checked = $bindable(false), class: className, children, ...rest }: Props = $props();
</script>

<label class={cn('flex cursor-pointer items-start gap-3 text-sm leading-relaxed text-foreground', className)}>
	<input type="checkbox" bind:checked class="mt-1 size-4 shrink-0 cursor-pointer accent-[var(--accent)]" {...rest} />
	<span>{@render children()}</span>
</label>
