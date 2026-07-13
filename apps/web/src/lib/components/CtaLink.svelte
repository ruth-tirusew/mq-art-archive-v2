<script lang="ts">
	import { cn } from '$lib/utils';

	type CtaVariant = 'primary' | 'secondary' | 'tertiary';
	type CtaTone = 'default' | 'on-dark';

	type Props = {
		href: string;
		variant?: CtaVariant;
		tone?: CtaTone;
		class?: string;
		children: import('svelte').Snippet;
	};

	let { href, variant = 'tertiary', tone = 'default', class: className, children }: Props = $props();

	const variantStyles: Record<CtaTone, Record<CtaVariant, string>> = {
		default: {
			primary:
				'rounded-full bg-foreground px-5 py-2.5 text-background shadow-sm hover:bg-accent hover:text-accent-foreground hover:shadow-md hover:-translate-y-0.5',
			secondary:
				'rounded-full border border-foreground/25 px-5 py-2.5 text-foreground hover:border-foreground hover:bg-foreground/5 hover:-translate-y-0.5',
			tertiary:
				'text-foreground underline decoration-accent decoration-2 underline-offset-8 hover:text-accent'
		},
		'on-dark': {
			primary:
				'rounded-full bg-cream px-5 py-2.5 text-ink shadow-sm hover:bg-accent hover:text-accent-foreground hover:shadow-md hover:-translate-y-0.5',
			secondary:
				'rounded-full border border-cream/30 px-5 py-2.5 text-cream hover:border-cream hover:bg-cream/10 hover:-translate-y-0.5',
			tertiary:
				'text-cream/80 underline decoration-accent decoration-2 underline-offset-8 hover:text-cream'
		}
	};
</script>

<a
	{href}
	class={cn(
		'inline-flex items-center gap-2 font-mono text-[11px] uppercase tracking-[0.2em] transition duration-200',
		variantStyles[tone][variant],
		className
	)}
>
	{@render children()}
</a>
