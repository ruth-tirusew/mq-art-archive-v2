<script lang="ts">
	import type { Snippet } from 'svelte';
	import CtaLink from '$lib/components/CtaLink.svelte';

	type Props = {
		eyebrow: string;
		id?: string;
		tone?: 'default' | 'on-dark';
		action?: { href: string; label: string };
		lead?: string;
		children: Snippet;
	};

	let { eyebrow, id, tone = 'default', action, lead, children }: Props = $props();
</script>

<header class="flex flex-wrap items-end justify-between gap-x-10 gap-y-6">
	<div class="flex max-w-3xl flex-col gap-4">
		<p class="t-eyebrow {tone === 'on-dark' ? 'text-ochre' : 'text-accent'}">{eyebrow}</p>
		<h2 {id} class="t-h2 {tone === 'on-dark' ? 'text-cream' : 'text-foreground'}">
			{@render children()}
		</h2>
		{#if lead}
			<p class="t-lead {tone === 'on-dark' ? 'text-cream/75' : 'text-muted-foreground'}">{lead}</p>
		{/if}
	</div>
	{#if action}
		<CtaLink href={action.href} {tone}>{action.label} <span aria-hidden="true">→</span></CtaLink>
	{/if}
</header>
