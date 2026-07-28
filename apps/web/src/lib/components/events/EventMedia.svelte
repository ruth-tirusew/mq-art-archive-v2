<script lang="ts">
	import type { Event } from '$lib/core/domain/events';

	let {
		event,
		class: className = ''
	}: {
		event: Event;
		class?: string;
	} = $props();

	let stage = $state<'primary' | 'fallback' | 'pattern'>('primary');

	const hues = [25, 40, 55, 200, 260, 320];
	const hue = $derived.by(() => {
		let h = 0;
		for (const ch of event.id || event.slug) h = (h * 31 + ch.charCodeAt(0)) >>> 0;
		return hues[h % hues.length];
	});

	const fallbackSrc = $derived(
		`https://picsum.photos/seed/${encodeURIComponent(event.slug || event.id)}/480/360`
	);

	$effect(() => {
		void event.id;
		stage = event.image_url ? 'primary' : 'fallback';
	});
</script>

{#if stage === 'primary' && event.image_url}
	<img
		src={event.image_url}
		alt=""
		class={className}
		loading="lazy"
		onerror={() => (stage = 'fallback')}
	/>
{:else if stage === 'fallback'}
	<img
		src={fallbackSrc}
		alt=""
		class={className}
		loading="lazy"
		onerror={() => (stage = 'pattern')}
	/>
{:else}
	<div
		class="flex items-end p-2 {className}"
		style="background-image: linear-gradient(145deg, oklch(0.88 0.04 {hue}), oklch(0.72 0.08 {hue}), oklch(0.45 0.06 {hue}))"
		aria-hidden="true"
	>
		<span class="font-mono text-[9px] uppercase tracking-[0.16em] text-background/80">
			{event.event_type}
		</span>
	</div>
{/if}
