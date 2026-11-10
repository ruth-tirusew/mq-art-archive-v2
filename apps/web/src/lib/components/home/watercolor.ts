/**
 * Timing and shape rules for the watercolor on the empty home page. Pure functions, so how a drop
 * looks and when the brush lays one down can be tested without a canvas.
 *
 * A drop is a pre-painted "stamp": many translucent, slightly offset washes of one irregular
 * outline, with specks of paper left showing through. Stamping it at the pointer, scaled up while
 * it spreads and faded as it dries, reads as pigment bleeding into wet paper.
 */

export interface Rgb {
	r: number;
	g: number;
	b: number;
}

export const STAMP_SIZE = 256;
export const WASH_LAYERS = 22;
export const WASH_ALPHA = 0.024;
export const GRAIN_SPECKS = 5200;

/** Milliseconds a drop stays visible before it has dried away. */
export const BLOOM_LIFE_MS = 6500;
export const MAX_BLOOMS = 65;

/** Small, fast, deterministic PRNG (mulberry32), so a seeded stamp always has the same shape. */
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

/** One uneven pool outline, as polar points around the stamp's centre. */
export function stampOutline(rng: () => number, points = 80): { angle: number; radius: number }[] {
	return Array.from({ length: points }, (_, i) => {
		const angle = (i / points) * Math.PI * 2;
		const radius = 74 + 13 * Math.sin(angle * 3 + 1) + 10 * Math.cos(angle * 7) + rng() * 12;
		return { angle, radius };
	});
}

/** Each wash is the same outline a touch larger, so the layers stack into soft, darker edges. */
export function washScale(layer: number): number {
	return 0.6 + layer * 0.017;
}

/** How big the drop is (as a fraction of its full size) and how strong it looks, `ageMs` after landing. */
export function bloomLook(ageMs: number): { spread: number; alpha: number } {
	const spread = 0.35 + 0.65 * (1 - Math.exp(-ageMs / 700));
	const wetting = Math.min(ageMs / 120, 1);
	const drying = Math.min((BLOOM_LIFE_MS - ageMs) / 2300, 1);
	return { spread, alpha: Math.max(0, wetting * drying * 0.8) };
}

export function isDry(ageMs: number): boolean {
	return ageMs >= BLOOM_LIFE_MS;
}

export interface BrushPoint {
	x: number;
	y: number;
	/** performance.now() when the point was seen. */
	t: number;
}

/** The brush lays a drop only after enough time and distance since the last one. */
export function shouldPaint(last: BrushPoint | null, next: BrushPoint): boolean {
	if (!last) return true;
	return next.t - last.t >= 55 && Math.hypot(next.x - last.x, next.y - last.y) >= 24;
}

/** The brush changes colour every few drops. */
export function pigmentIndex(drops: number, pigments: number): number {
	return Math.floor(drops / 5) % pigments;
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
