<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '../cn';

	type Props = Omit<HTMLAttributes<HTMLDivElement>, 'children'> & {
		/** default: paper card. outline: hairline only. ink: dark feature card (use tone on-dark inside). */
		variant?: 'default' | 'outline' | 'ink';
		padded?: boolean;
		class?: string;
		children: Snippet;
	};

	let { variant = 'default', padded = true, class: className, children, ...rest }: Props = $props();

	const variants = {
		default: 'bg-card text-card-foreground border border-border/60',
		outline: 'border border-border',
		ink: 'bg-ink text-cream'
	};
</script>

<div class={cn('rounded-sm', variants[variant], padded && 'p-6 md:p-8', className)} {...rest}>
	{@render children()}
</div>
