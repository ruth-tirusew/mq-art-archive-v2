import { describe, expect, it } from 'vitest';
import {
	BLOOM_LIFE_MS,
	bloomLook,
	createRng,
	isDry,
	pigmentIndex,
	shouldPaint,
	stampOutline,
	washScale
} from './watercolor';

describe('bloomLook', () => {
	it('starts small and spreads quickly toward full size', () => {
		expect(bloomLook(0).spread).toBeCloseTo(0.35);
		expect(bloomLook(700).spread).toBeGreaterThan(0.75);
		expect(bloomLook(5000).spread).toBeLessThanOrEqual(1);
	});

	it('never shrinks while it spreads', () => {
		let last = 0;
		for (let t = 0; t <= BLOOM_LIFE_MS; t += 50) {
			expect(bloomLook(t).spread).toBeGreaterThanOrEqual(last);
			last = bloomLook(t).spread;
		}
	});

	it('wets in fast, holds, then dries to nothing', () => {
		expect(bloomLook(0).alpha).toBe(0);
		expect(bloomLook(120).alpha).toBeCloseTo(0.8);
		expect(bloomLook(2000).alpha).toBeCloseTo(0.8);
		expect(bloomLook(5500).alpha).toBeLessThan(0.8);
		expect(bloomLook(5500).alpha).toBeGreaterThan(0);
		expect(bloomLook(BLOOM_LIFE_MS).alpha).toBe(0);
		expect(bloomLook(BLOOM_LIFE_MS + 1000).alpha).toBe(0);
	});

	it('is dry exactly when its life is over', () => {
		expect(isDry(BLOOM_LIFE_MS - 1)).toBe(false);
		expect(isDry(BLOOM_LIFE_MS)).toBe(true);
	});
});

describe('stampOutline', () => {
	it('is deterministic for a seed and different for another', () => {
		expect(stampOutline(createRng(3))).toEqual(stampOutline(createRng(3)));
		expect(stampOutline(createRng(3))).not.toEqual(stampOutline(createRng(4)));
	});

	it('stays one lobed blob inside the stamp, never a star or a speck', () => {
		for (let seed = 1; seed < 50; seed++) {
			for (const { radius } of stampOutline(createRng(seed))) {
				expect(radius).toBeGreaterThan(50);
				expect(radius).toBeLessThan(110);
			}
		}
	});

	it('goes once around the circle', () => {
		const pts = stampOutline(createRng(1), 80);
		expect(pts).toHaveLength(80);
		expect(pts[0].angle).toBe(0);
		expect(pts[79].angle).toBeLessThan(Math.PI * 2);
	});

	it('has washes that grow from 60% of the outline, so edges stack up darker', () => {
		expect(washScale(0)).toBeCloseTo(0.6);
		expect(washScale(21)).toBeGreaterThan(washScale(20));
		expect(washScale(21)).toBeLessThan(1);
	});
});

describe('brush', () => {
	it('lays the first drop at once', () => {
		expect(shouldPaint(null, { x: 0, y: 0, t: 0 })).toBe(true);
	});

	it('waits for both enough time and enough distance', () => {
		const last = { x: 0, y: 0, t: 1000 };
		expect(shouldPaint(last, { x: 100, y: 0, t: 1020 })).toBe(false); // too soon
		expect(shouldPaint(last, { x: 5, y: 5, t: 1100 })).toBe(false); // too close
		expect(shouldPaint(last, { x: 30, y: 10, t: 1060 })).toBe(true);
	});

	it('changes colour every five drops and wraps around the set', () => {
		expect([0, 4, 5, 9, 10].map((n) => pigmentIndex(n, 3))).toEqual([0, 0, 1, 1, 2]);
		expect(pigmentIndex(15, 3)).toBe(0);
	});
});
