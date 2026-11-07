/**
 * Geometry and timing for the watercolor spill on the empty home page. Pure functions, so the
 * look (how fast a drop spreads, how its edge wobbles, when it dries) is testable without a canvas.
 */

export interface Rgb {
	r: number;
	g: number;
	b: number;
}

export interface Bloom {
	x: number;
	y: number;
	/** Radius in px when fully spread. */
	radius: number;
	pigment: Rgb;
	/** Seconds on the clock when the drop landed. */
	born: number;
	/** Per-bloom wobble: harmonic amplitudes and phases for the edge, and how fast each one drifts. */
	harmonics: { k: number; amp: number; phase: number; drift: number }[];
}

/** Seconds a drop stays visible: it spills fast, creeps for a while, then dries away. */
export const BLOOM_LIFE = 18;
const DRY_FROM = 4;

/** Small, fast, deterministic PRNG (mulberry32), so a seeded bloom always has the same shape. */
export function createRng(seed: number): () => number {
	let a = seed >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

export function createBloom(x: number, y: number, radius: number, pigment: Rgb, born: number, rng: () => number): Bloom {
	// Low harmonics make a lobed, uneven pool; higher ones a fine ragged edge. Amplitude falls
	// off with k so the shape stays one blob rather than a star.
	const harmonics = [2, 3, 4, 5, 7].map((k) => ({
		k,
		amp: (0.16 / Math.sqrt(k - 0.5)) * (0.55 + rng() * 0.9),
		phase: rng() * Math.PI * 2,
		drift: (rng() - 0.5) * 0.5
	}));
	return { x, y, radius, pigment, born, harmonics };
}

/** How far the drop has spread, 0..1: most of it in the first second, then a slow creep. */
export function spread(age: number): number {
	if (age <= 0) return 0;
	return 0.82 * (1 - Math.exp(-age / 0.9)) + 0.18 * (1 - Math.exp(-age / 7));
}

/** Visible strength, 0..1: wet quickly, holds, then dries to nothing at BLOOM_LIFE. */
export function wetness(age: number): number {
	if (age <= 0 || age >= BLOOM_LIFE) return 0;
	const attack = Math.min(1, age / 0.35);
	const drying = age < DRY_FROM ? 1 : 1 - (age - DRY_FROM) / (BLOOM_LIFE - DRY_FROM);
	return attack * Math.pow(Math.max(0, drying), 1.5);
}

export function isDry(bloom: Bloom, now: number): boolean {
	return now - bloom.born >= BLOOM_LIFE;
}

/** The outline of the pool at `now`, as points around its centre. */
export function outline(bloom: Bloom, now: number, steps = 56): { x: number; y: number }[] {
	const age = now - bloom.born;
	const base = bloom.radius * spread(age);
	const pts: { x: number; y: number }[] = [];
	for (let i = 0; i < steps; i++) {
		const angle = (i / steps) * Math.PI * 2;
		let wobble = 1;
		for (const h of bloom.harmonics) wobble += h.amp * Math.sin(h.k * angle + h.phase + h.drift * age);
		const r = Math.max(0, base * wobble);
		pts.push({ x: bloom.x + Math.cos(angle) * r, y: bloom.y + Math.sin(angle) * r });
	}
	return pts;
}

/** A new drop is only laid down once the brush has moved far enough from the last one. */
export function shouldDrop(last: { x: number; y: number } | null, next: { x: number; y: number }, spacing: number): boolean {
	if (!last) return true;
	return Math.hypot(next.x - last.x, next.y - last.y) >= spacing;
}

/** Faster strokes carry more water: a bigger, thinner pool. Clamped so it never gets absurd. */
export function dropRadius(speedPxPerMs: number, min = 70, max = 150): number {
	return Math.min(max, min + speedPxPerMs * 55);
}

/** Resolve any CSS colour (including oklch tokens) to sRGB by letting the browser paint it. */
export function resolveCssColor(css: string, ctx: CanvasRenderingContext2D): Rgb | null {
	ctx.clearRect(0, 0, 1, 1);
	ctx.fillStyle = '#000';
	ctx.fillStyle = css; // an invalid value leaves the previous (black) fill, so test for it below
	ctx.fillRect(0, 0, 1, 1);
	const [r, g, b, a] = ctx.getImageData(0, 0, 1, 1).data;
	return a === 0 ? null : { r, g, b };
}
