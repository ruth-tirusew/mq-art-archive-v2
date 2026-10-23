<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { HTMLButtonAttributes } from 'svelte/elements';
	import { cn } from '../cn';

	type Variant = 'primary' | 'accent' | 'secondary' | 'ghost' | 'destructive' | 'link';
	type Tone = 'default' | 'on-dark';
	type Size = 'sm' | 'md';

	type Props = Omit<HTMLButtonAttributes, 'children'> & {
		variant?: Variant;
		tone?: Tone;
		size?: Size;
		/** Renders an anchor instead of a button. */
		href?: string;
		/** Shows a spinner and blocks interaction. Use for in-flight submits. */
		loading?: boolean;
		class?: string;
		children: Snippet;
	};

	let {
		variant = 'primary',
		tone = 'default',
		size = 'md',
		href,
		loading = false,
		disabled = false,
		type = 'button',
		class: className,
		children,
		...rest
	}: Props = $props();

	const styles: Record<Tone, Record<Variant, string>> = {
		default: {
			primary: 'rounded-full bg-foreground text-background shadow-sm hover:bg-accent hover:text-accent-foreground hover:shadow-md',
			accent: 'rounded-full bg-accent text-accent-foreground shadow-sm hover:bg-foreground hover:text-background hover:shadow-md',
			secondary: 'rounded-full border border-foreground/25 text-foreground hover:border-foreground hover:bg-foreground/5',
			ghost: 'rounded-full text-foreground hover:bg-foreground/5',
			destructive: 'rounded-full bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/85',
			link: 'border-b-2 border-accent pb-1.5 text-foreground hover:text-accent'
		},
		'on-dark': {
			primary: 'rounded-full bg-cream text-ink shadow-sm hover:bg-accent hover:text-accent-foreground hover:shadow-md',
			accent: 'rounded-full bg-accent text-accent-foreground shadow-sm hover:bg-cream hover:text-ink hover:shadow-md',
			secondary: 'rounded-full border border-cream/30 text-cream hover:border-cream hover:bg-cream/10',
			ghost: 'rounded-full text-cream hover:bg-cream/10',
			destructive: 'rounded-full bg-destructive text-destructive-foreground hover:bg-destructive/85',
			link: 'border-b-2 border-accent pb-1.5 text-cream/80 hover:text-cream'
		}
	};

	const sizes: Record<Size, string> = { sm: 'px-3.5 py-1.5', md: 'px-5 py-2.5' };

	const inert = $derived(disabled || loading);
	const classes = $derived(
		cn(
			't-eyebrow inline-flex items-center justify-center gap-2 transition duration-200',
			'disabled:pointer-events-none disabled:opacity-50 aria-disabled:pointer-events-none aria-disabled:opacity-50',
			variant !== 'link' && sizes[size],
			variant !== 'link' && variant !== 'ghost' && 'hover:-translate-y-0.5',
			styles[tone][variant],
			className
		)
	);
</script>

{#if href}
	<a {href} class={classes} aria-disabled={inert || undefined} tabindex={inert ? -1 : undefined}>
		{@render children()}
	</a>
{:else}
	<button {type} class={classes} disabled={inert} aria-busy={loading || undefined} {...rest}>
		{#if loading}
			<span class="size-3 animate-spin rounded-full border-2 border-current border-t-transparent" aria-hidden="true"></span>
		{/if}
		{@render children()}
	</button>
{/if}
