<script lang="ts">
	import { onMount } from 'svelte';
	import {
		BLOOM_LIFE_MS,
		GRAIN_SPECKS,
		MAX_BLOOMS,
		STAMP_SIZE,
		WASH_ALPHA,
		WASH_LAYERS,
		bloomLook,
		createRng,
		isDry,
		pigmentIndex,
		resolveCssColor,
		shouldPaint,
		stampOutline,
		washScale,
		type BrushPoint,
		type Rgb
	} from './watercolor';

	type Bloom = { x: number; y: number; born: number; size: number; stamp: HTMLCanvasElement };

	// The hero defines these from the site palette (see packages/ui theme.css), so paints stay on-brand.
	const PIGMENT_TOKENS = ['--pigment-red', '--pigment-ochre', '--pigment-olive', '--pigment-indigo'];
	const IDLE_AFTER_MS = 5000;

	let canvas: HTMLCanvasElement;

	/** A pigment stamp: irregular translucent washes, with specks of paper left showing through. */
	function createStamp({ r, g, b }: Rgb, rng: () => number): HTMLCanvasElement {
		const stamp = document.createElement('canvas');
		stamp.width = stamp.height = STAMP_SIZE;
		const ctx = stamp.getContext('2d');
		if (!ctx) return stamp;
		const centre = STAMP_SIZE / 2;
		const outline = stampOutline(rng);
		ctx.fillStyle = `rgb(${r},${g},${b})`;
		for (let layer = 0; layer < WASH_LAYERS; layer++) {
			const scale = washScale(layer);
			ctx.globalAlpha = WASH_ALPHA;
			ctx.beginPath();
			outline.forEach(({ angle, radius }, i) => {
				const rr = radius * scale + rng() * 7;
				const x = centre + Math.cos(angle) * rr;
				const y = centre + Math.sin(angle) * rr;
				if (i === 0) ctx.moveTo(x, y);
				else ctx.lineTo(x, y);
			});
			ctx.closePath();
			ctx.fill();
		}
		// Tiny gaps in the pigment leave the paper grain visible.
		ctx.globalCompositeOperation = 'destination-out';
		for (let i = 0; i < GRAIN_SPECKS; i++) {
			ctx.globalAlpha = rng() * 0.23;
			ctx.fillRect(rng() * STAMP_SIZE, rng() * STAMP_SIZE, 1, 1);
		}
		return stamp;
	}

	onMount(() => {
		// The canvas never takes pointer events itself; its parent (the hero) is what paints.
		// (A prop would arrive too late: children mount before the parent's bind:this is set.)
		const parent = canvas.parentElement;
		const ctx = canvas.getContext('2d');
		const probe = document.createElement('canvas');
		probe.width = probe.height = 1;
		const probeCtx = probe.getContext('2d', { willReadFrequently: true });
		if (!parent || !ctx || !probeCtx) return;

		const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
		const styles = getComputedStyle(parent);
		const rng = createRng((Date.now() & 0xffff) + 1);
		const stamps = PIGMENT_TOKENS.map((token) => resolveCssColor(styles.getPropertyValue(token).trim(), probeCtx))
			.filter((c): c is Rgb => c !== null)
			.map((c) => createStamp(c, rng));
		if (stamps.length === 0) return;

		let width = 0;
		let height = 0;
		let frame = 0;
		let last: BrushPoint | null = null;
		let drops = 0;
		let blooms: Bloom[] = [];
		let visible = true;
		let lastInput = performance.now();
		let idleTimer: ReturnType<typeof setTimeout> | undefined;

		function resize() {
			const rect = parent!.getBoundingClientRect();
			width = rect.width;
			height = rect.height;
			const dpr = Math.min(window.devicePixelRatio || 1, 2);
			canvas.width = Math.round(width * dpr);
			canvas.height = Math.round(height * dpr);
			ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
		}

		function draw(time: number) {
			ctx!.clearRect(0, 0, width, height);
			blooms = blooms.filter((b) => !isDry(time - b.born));
			for (const bloom of blooms) {
				const { spread, alpha } = bloomLook(time - bloom.born);
				const size = bloom.size * spread;
				ctx!.globalAlpha = alpha;
				ctx!.drawImage(bloom.stamp, bloom.x - size / 2, bloom.y - size / 2, size, size);
			}
			frame = blooms.length ? requestAnimationFrame(draw) : 0;
		}

		function drop(x: number, y: number, now: number, size = 210 + rng() * 130) {
			blooms.push({ x, y, born: now, size, stamp: stamps[pigmentIndex(drops, stamps.length)] });
			if (blooms.length > MAX_BLOOMS) blooms.shift();
			drops++;
			if (!frame) frame = requestAnimationFrame(draw);
		}

		function paint(event: PointerEvent) {
			if (reducedMotion.matches) return;
			lastInput = performance.now();
			const rect = parent!.getBoundingClientRect();
			const point = { x: event.clientX - rect.left, y: event.clientY - rect.top, t: lastInput };
			if (!shouldPaint(last, point)) return;
			last = point;
			drop(point.x, point.y, point.t);
		}

		function leave() {
			last = null;
		}

		// Without a pointer (touch devices, or someone just reading) the page still breathes: now
		// and then a drop lands by itself, but only while the hero is on screen and the tab is open.
		function ambient(delay: number) {
			idleTimer = setTimeout(() => {
				const now = performance.now();
				if (visible && !document.hidden && !reducedMotion.matches && now - lastInput > IDLE_AFTER_MS) {
					drop(width * (0.15 + rng() * 0.7), height * (0.15 + rng() * 0.7), now);
				}
				ambient(4500 + rng() * 3000);
			}, delay);
		}

		const resizeObserver = new ResizeObserver(() => {
			resize();
			if (blooms.length && !frame) frame = requestAnimationFrame(draw);
		});
		resizeObserver.observe(parent);
		const visibility = new IntersectionObserver(([entry]) => (visible = entry.isIntersecting));
		visibility.observe(parent);

		parent.addEventListener('pointermove', paint);
		parent.addEventListener('pointerdown', paint);
		parent.addEventListener('pointerleave', leave);

		resize();
		ambient(600); // a first drop soon after load, so the effect announces itself

		return () => {
			clearTimeout(idleTimer);
			cancelAnimationFrame(frame);
			resizeObserver.disconnect();
			visibility.disconnect();
			parent.removeEventListener('pointermove', paint);
			parent.removeEventListener('pointerdown', paint);
			parent.removeEventListener('pointerleave', leave);
		};
	});
</script>

<!-- Multiply lets overlapping pigments darken and mix, like washes on paper. -->
<canvas
	bind:this={canvas}
	aria-hidden="true"
	class="pointer-events-none absolute inset-0 z-10 h-full w-full mix-blend-multiply dark:mix-blend-screen"
></canvas>
