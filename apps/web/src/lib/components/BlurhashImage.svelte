<script lang="ts">
	import { decode } from 'blurhash';
	import { deliverySrcset, deliveryUrl } from '$lib/utils/cloudinary';

	type Props = {
		src: string;
		alt: string;
		/** Compact placeholder for this image; without one the image just loads in. */
		blurhash?: string;
		/** The widest this image is shown, in CSS pixels. Cloudinary images are delivered at up to 2x this. */
		width?: number;
		sizes?: string;
		loading?: 'lazy' | 'eager';
		/** Classes for the <img>, e.g. object-fit and hover effects. */
		class?: string;
		onerror?: () => void;
	};

	let { src, alt, blurhash = '', width, sizes, loading = 'lazy', class: className = '', onerror }: Props = $props();

	let loaded = $state(false);
	let failed = $state(false);
	let canvas = $state<HTMLCanvasElement>();
	let img = $state<HTMLImageElement>();

	const delivered = $derived(deliveryUrl(src, width ? width * 2 : undefined));
	const srcset = $derived(width ? undefined : deliverySrcset(src, [400, 800, 1200, 1600]));

	// A new image starts blurred again.
	$effect(() => {
		src;
		loaded = false;
		failed = false;
	});

	// Paint the placeholder once the canvas exists. 32x32 is plenty for something that is blurred.
	$effect(() => {
		if (!canvas || !blurhash) return;
		try {
			const pixels = decode(blurhash, 32, 32);
			const ctx = canvas.getContext('2d');
			if (!ctx) return;
			const data = ctx.createImageData(32, 32);
			data.data.set(pixels);
			ctx.putImageData(data, 0, 0);
		} catch {
			// A malformed hash just means no placeholder.
		}
	});

	// An image that was already cached (or server-rendered) can finish before hydration, so the
	// load event never reaches us. Check for it.
	$effect(() => {
		if (img?.complete && img.naturalWidth > 0) loaded = true;
	});
</script>

<div class="relative h-full w-full overflow-hidden">
	{#if blurhash}
		<canvas
			bind:this={canvas}
			width="32"
			height="32"
			aria-hidden="true"
			class="absolute inset-0 h-full w-full scale-110 object-cover transition-opacity duration-500 {loaded ? 'opacity-0' : 'opacity-100'}"
		></canvas>
	{/if}
	{#if !failed}
		<img
			bind:this={img}
			src={delivered}
			{srcset}
			{sizes}
			{alt}
			{loading}
			decoding="async"
			onload={() => (loaded = true)}
			onerror={() => {
				failed = true;
				onerror?.();
			}}
			class="{className} transition-opacity duration-500 {blurhash && !loaded ? 'opacity-0' : 'opacity-100'}"
		/>
	{/if}
</div>
