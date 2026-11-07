<script lang="ts">
	import { onMount } from 'svelte';
	import {
		createBloom,
		createRng,
		dropRadius,
		isDry,
		outline,
		shouldDrop,
		resolveCssColor,
		wetness,
		type Bloom,
		type Rgb
	} from './watercolor';

	// Palette tokens, so the paints are always the site's colours (see packages/ui theme.css).
	const PIGMENT_TOKENS = ['--accent', '--ochre', '--indigo-deep', '--terracotta', '--success'];
	const MAX_BLOOMS = 36;
	const SPACING = 56;
	const IDLE_AFTER_MS = 5000;

	let canvas: HTMLCanvasElement;

	onMount(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		// The canvas never takes pointer events itself; its parent (the hero) is what paints.
		// (A prop would arrive too late: children mount before the parent's bind:this is set.)
		const target = canvas.parentElement;
		if (!target) return;
		const ctx = canvas.getContext('2d', { willReadFrequently: false });
		const probe = document.createElement('canvas');
		probe.width = probe.height = 1;
		const probeCtx = probe.getContext('2d', { willReadFrequently: true });
		if (!ctx || !probeCtx) return;

		const styles = getComputedStyle(document.documentElement);
		const pigments = PIGMENT_TOKENS.map((t) => resolveCssColor(styles.getPropertyValue(t).trim(), probeCtx)).filter(
			(c): c is Rgb => c !== null
		);
		if (pigments.length === 0) return;

		const rng = createRng((Date.now() & 0xffff) + 1);
		const clock = () => performance.now() / 1000;
		let blooms: Bloom[] = [];
		let width = 0;
		let height = 0;
		let frame = 0;
		let pigmentIndex = Math.floor(rng() * pigments.length);
		let last: { x: number; y: number } | null = null;
		let lastMove = { t: 0, x: 0, y: 0 };
		let lastInput = performance.now();
		let visible = true;
		let idleTimer: ReturnType<typeof setTimeout> | undefined;

		function resize() {
			const rect = canvas.getBoundingClientRect();
			const dpr = Math.min(window.devicePixelRatio || 1, 2);
			width = rect.width;
			height = rect.height;
			canvas.width = Math.round(width * dpr);
			canvas.height = Math.round(height * dpr);
			ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
			draw();
		}

		function drop(x: number, y: number, radius: number) {
			blooms.push(createBloom(x, y, radius, pigments[pigmentIndex], clock(), rng));
			if (blooms.length > MAX_BLOOMS) blooms.shift();
			if (!frame) frame = requestAnimationFrame(tick);
		}

		function draw() {
			const c = ctx!;
			c.clearRect(0, 0, width, height);
			const now = clock();
			for (const bloom of blooms) {
				const wet = wetness(now - bloom.born);
				if (wet <= 0) continue;
				const pts = outline(bloom, now);
				const { r, g, b } = bloom.pigment;
				const reach = bloom.radius * 1.15;
				c.beginPath();
				pts.forEach((p, i) => (i === 0 ? c.moveTo(p.x, p.y) : c.lineTo(p.x, p.y)));
				c.closePath();
				// Pigment pools at the rim and thins toward the middle: that is the watercolor look.
				const wash = c.createRadialGradient(bloom.x, bloom.y, reach * 0.1, bloom.x, bloom.y, reach);
				wash.addColorStop(0, `rgba(${r},${g},${b},${0.07 * wet})`);
				wash.addColorStop(0.7, `rgba(${r},${g},${b},${0.14 * wet})`);
				wash.addColorStop(1, `rgba(${r},${g},${b},${0.22 * wet})`);
				c.fillStyle = wash;
				c.fill();
				c.lineWidth = 1;
				c.strokeStyle = `rgba(${r},${g},${b},${0.16 * wet})`;
				c.stroke();
			}
		}

		function tick() {
			frame = 0;
			const now = clock();
			blooms = blooms.filter((b) => !isDry(b, now));
			draw();
			if (blooms.length > 0) frame = requestAnimationFrame(tick);
		}

		function local(event: PointerEvent) {
			const rect = canvas.getBoundingClientRect();
			return { x: event.clientX - rect.left, y: event.clientY - rect.top };
		}

		function onMove(event: PointerEvent) {
			lastInput = performance.now();
			const p = local(event);
			const dt = Math.max(1, performance.now() - lastMove.t);
			const speed = Math.hypot(p.x - lastMove.x, p.y - lastMove.y) / dt;
			lastMove = { t: performance.now(), x: p.x, y: p.y };
			if (!shouldDrop(last, p, SPACING)) return;
			last = p;
			drop(p.x, p.y, dropRadius(speed));
		}

		function onEnter() {
			// A new dip of the brush: new colour, and the first touch lays a drop straight away.
			pigmentIndex = (pigmentIndex + 1 + Math.floor(rng() * (pigments.length - 1))) % pigments.length;
			last = null;
		}

		function onDown(event: PointerEvent) {
			lastInput = performance.now();
			const p = local(event);
			last = p;
			drop(p.x, p.y, 130);
		}

		// Without a pointer (touch devices, or someone just reading) the page still breathes: now
		// and then a drop lands by itself, but only while the hero is on screen and the tab is open.
		function scheduleIdle() {
			idleTimer = setTimeout(
				() => {
					if (visible && !document.hidden && performance.now() - lastInput > IDLE_AFTER_MS) {
						pigmentIndex = Math.floor(rng() * pigments.length);
						drop(width * (0.15 + rng() * 0.7), height * (0.15 + rng() * 0.7), 90 + rng() * 60);
					}
					scheduleIdle();
				},
				4500 + rng() * 3000
			);
		}

		const resizeObserver = new ResizeObserver(resize);
		resizeObserver.observe(canvas);
		const visibility = new IntersectionObserver(([entry]) => (visible = entry.isIntersecting));
		visibility.observe(canvas);

		target.addEventListener('pointermove', onMove);
		target.addEventListener('pointerenter', onEnter);
		target.addEventListener('pointerdown', onDown);

		resize();
		// A first drop so the effect announces itself before anyone moves.
		const intro = setTimeout(() => drop(width * 0.68, height * 0.38, 120), 500);
		scheduleIdle();

		return () => {
			clearTimeout(intro);
			clearTimeout(idleTimer);
			cancelAnimationFrame(frame);
			resizeObserver.disconnect();
			visibility.disconnect();
			target.removeEventListener('pointermove', onMove);
			target.removeEventListener('pointerenter', onEnter);
			target.removeEventListener('pointerdown', onDown);
		};
	});
</script>

<!-- Multiply lets overlapping pigments darken and mix, like washes on paper. -->
<canvas
	bind:this={canvas}
	aria-hidden="true"
	class="pointer-events-none absolute inset-0 z-10 h-full w-full mix-blend-multiply dark:mix-blend-screen"
></canvas>
