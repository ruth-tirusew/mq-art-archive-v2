<script lang="ts">
	import { onDestroy } from 'svelte';
	import { cn } from '$lib/utils';

	type Props = {
		/** The address to copy. */
		url: string;
		label?: string;
		class?: string;
	};

	let { url, label = 'Copy link', class: className }: Props = $props();

	let copied = $state(false);
	let timer: ReturnType<typeof setTimeout> | undefined;
	onDestroy(() => clearTimeout(timer));

	async function copy() {
		let ok = false;
		try {
			await navigator.clipboard.writeText(url);
			ok = true;
		} catch {
			// Older browsers, or a page without clipboard permission: fall back to selecting a hidden field.
			try {
				const field = document.createElement('textarea');
				field.value = url;
				field.setAttribute('readonly', '');
				field.style.cssText = 'position:fixed;opacity:0;pointer-events:none';
				document.body.appendChild(field);
				field.select();
				ok = document.execCommand('copy');
				field.remove();
			} catch {
				ok = false;
			}
		}
		if (!ok) {
			// Last resort: show the address so it can be copied by hand.
			window.prompt('Copy this link', url);
			return;
		}
		copied = true;
		clearTimeout(timer);
		timer = setTimeout(() => (copied = false), 2000);
	}
</script>

<button
	type="button"
	onclick={copy}
	class={cn(
		'inline-flex items-center gap-1.5 rounded-sm font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground transition hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40',
		copied && 'text-accent',
		className
	)}
>
	<svg viewBox="0 0 24 24" class="h-3.5 w-3.5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
		{#if copied}
			<path d="M20 6 9 17l-5-5" />
		{:else}
			<rect width="14" height="14" x="8" y="8" rx="2" />
			<path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
		{/if}
	</svg>
	{copied ? 'Link copied' : label}
	<span class="sr-only" role="status" aria-live="polite">{copied ? 'Link copied to the clipboard' : ''}</span>
</button>
