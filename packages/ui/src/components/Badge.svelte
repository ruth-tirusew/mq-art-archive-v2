<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '../cn';

	type Tone = 'neutral' | 'accent' | 'success' | 'warning' | 'destructive' | 'ink';

	type Props = { tone?: Tone; class?: string; children: Snippet };

	let { tone = 'neutral', class: className, children }: Props = $props();

	const tones: Record<Tone, string> = {
		neutral: 'border-border bg-muted text-muted-foreground',
		accent: 'border-accent/30 bg-accent/10 text-accent',
		success: 'border-success/30 bg-success/10 text-success',
		warning: 'border-warning/30 bg-warning/10 text-warning',
		destructive: 'border-destructive/30 bg-destructive/10 text-destructive',
		ink: 'border-ink bg-ink text-cream'
	};
</script>

<span class={cn('t-meta inline-flex items-center rounded-full border px-2.5 py-0.5 text-[10px]', tones[tone], className)}>
	{@render children()}
</span>
