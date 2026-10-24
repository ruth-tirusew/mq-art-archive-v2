<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '../cn';

	type Tone = 'info' | 'success' | 'warning' | 'error';

	type Props = {
		tone?: Tone;
		title?: string;
		/** Adds a dismiss button; the parent clears its own message in the callback. */
		ondismiss?: () => void;
		class?: string;
		children: Snippet;
	};

	let { tone = 'info', title, ondismiss, class: className, children }: Props = $props();

	const tones: Record<Tone, string> = {
		info: 'border-border bg-card text-foreground',
		success: 'border-success/30 bg-success/10 text-success',
		warning: 'border-warning/30 bg-warning/10 text-warning',
		error: 'border-destructive/30 bg-destructive/10 text-destructive'
	};
</script>

<div
	role={tone === 'error' ? 'alert' : 'status'}
	class={cn('flex items-start justify-between gap-4 rounded-sm border px-4 py-3 text-sm', tones[tone], className)}
>
	<div class="flex flex-col gap-0.5">
		{#if title}<p class="t-eyebrow">{title}</p>{/if}
		<div class="leading-relaxed">{@render children()}</div>
	</div>
	{#if ondismiss}
		<button type="button" onclick={ondismiss} aria-label="Dismiss" class="-mr-1 shrink-0 px-1 text-lg leading-none opacity-60 hover:opacity-100">×</button>
	{/if}
</div>
