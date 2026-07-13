<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';

	type Props = {
		eyebrow: string;
		title: string;
		body: string;
		ctaLabel: string;
		ctaHref: string;
		class?: string;
	};

	let { eyebrow, title, body, ctaLabel, ctaHref, class: className = '' }: Props = $props();
</script>

<div
	class="rounded-sm border border-dashed border-border/70 bg-card/30 p-8 md:p-12 {className}"
>
	<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">{eyebrow}</p>
	<h3 class="mt-4 max-w-lg font-display text-2xl leading-tight text-foreground md:text-3xl">
		{title}
	</h3>
	<p class="mt-4 max-w-xl text-sm leading-relaxed text-muted-foreground">{body}</p>
	<CtaLink href={ctaHref} variant="secondary" class="mt-6">{ctaLabel} →</CtaLink>
</div>
