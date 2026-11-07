import { describe, expect, it } from 'vitest';
import {
	BLOOM_LIFE,
	createBloom,
	createRng,
	dropRadius,
	isDry,
	outline,
	shouldDrop,
	spread,
	wetness
} from './watercolor';

const pigment = { r: 139, g: 40, b: 30 };

describe('spread and wetness', () => {
	it('spills fast then creeps: most of the spread happens in the first second', () => {
		expect(spread(0)).toBe(0);
		expect(spread(1)).toBeGreaterThan(0.55);
		expect(spread(1)).toBeLessThan(spread(6));
		expect(spread(60)).toBeLessThanOrEqual(1);
	});

	it('never shrinks while it is spreading', () => {
		let last = 0;
		for (let t = 0; t <= BLOOM_LIFE; t += 0.25) {
			expect(spread(t)).toBeGreaterThanOrEqual(last);
			last = spread(t);
		}
	});

	it('wets in quickly, holds, then dries to nothing', () => {
		expect(wetness(0)).toBe(0);
		expect(wetness(0.35)).toBe(1);
		expect(wetness(3)).toBe(1);
		expect(wetness(BLOOM_LIFE / 2)).toBeLessThan(1);
		expect(wetness(BLOOM_LIFE / 2)).toBeGreaterThan(0);
		expect(wetness(BLOOM_LIFE)).toBe(0);
	});
});

describe('bloom outline', () => {
	it('is deterministic for a seed and a different shape for another', () => {
		const a = createBloom(100, 100, 100, pigment, 0, createRng(7));
		const b = createBloom(100, 100, 100, pigment, 0, createRng(7));
		const c = createBloom(100, 100, 100, pigment, 0, createRng(8));
		expect(outline(a, 3)).toEqual(outline(b, 3));
		expect(outline(a, 3)).not.toEqual(outline(c, 3));
	});

	it('stays one blob: radius is positive and within a sane band of the nominal radius', () => {
		for (let seed = 1; seed < 40; seed++) {
			const bloom = createBloom(0, 0, 100, pigment, 0, createRng(seed));
			for (const p of outline(bloom, BLOOM_LIFE - 1)) {
				const r = Math.hypot(p.x, p.y);
				expect(r).toBeGreaterThan(30);
				expect(r).toBeLessThan(190);
			}
		}
	});

	it('is a point at the moment the drop lands', () => {
		const bloom = createBloom(50, 60, 100, pigment, 5, createRng(1));
		for (const p of outline(bloom, 5)) expect(p).toEqual({ x: 50, y: 60 });
	});

	it('is dry once its life is over', () => {
		const bloom = createBloom(0, 0, 100, pigment, 10, createRng(1));
		expect(isDry(bloom, 10 + BLOOM_LIFE - 0.1)).toBe(false);
		expect(isDry(bloom, 10 + BLOOM_LIFE)).toBe(true);
	});
});

describe('brush', () => {
	it('lays the first drop at once and later ones only after enough movement', () => {
		expect(shouldDrop(null, { x: 0, y: 0 }, 40)).toBe(true);
		expect(shouldDrop({ x: 0, y: 0 }, { x: 20, y: 10 }, 40)).toBe(false);
		expect(shouldDrop({ x: 0, y: 0 }, { x: 30, y: 30 }, 40)).toBe(true);
	});

	it('gives faster strokes bigger pools, within limits', () => {
		expect(dropRadius(0)).toBe(70);
		expect(dropRadius(1)).toBeGreaterThan(dropRadius(0.2));
		expect(dropRadius(50)).toBe(150);
	});
});
