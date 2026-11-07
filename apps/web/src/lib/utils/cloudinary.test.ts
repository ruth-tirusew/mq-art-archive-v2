import { describe, expect, it } from 'vitest';
import { deliverySrcset, deliveryUrl, isCloudinaryUpload } from './cloudinary';

const original = 'https://res.cloudinary.com/demo/image/upload/v1699/mq/abc.jpg';

describe('deliveryUrl', () => {
	it('asks Cloudinary for lossless WebP at the width needed', () => {
		expect(deliveryUrl(original, 800)).toBe('https://res.cloudinary.com/demo/image/upload/c_limit,w_800,f_webp,q_100/v1699/mq/abc.jpg');
	});

	it('keeps the original size when no width is given', () => {
		expect(deliveryUrl(original)).toBe('https://res.cloudinary.com/demo/image/upload/f_webp,q_100/v1699/mq/abc.jpg');
	});

	it('leaves images that are not in Cloudinary exactly as they are', () => {
		const unsplash = 'https://images.unsplash.com/photo-1?w=1200&q=80';
		expect(deliveryUrl(unsplash, 800)).toBe(unsplash);
		expect(deliveryUrl('/local/a.jpg', 800)).toBe('/local/a.jpg');
	});

	it('does not stack a second transformation on one already applied', () => {
		const done = 'https://res.cloudinary.com/demo/image/upload/c_limit,w_400,f_webp,q_100/mq/abc.jpg';
		expect(deliveryUrl(done, 800)).toBe(done);
	});

	it('only treats https Cloudinary image uploads as Cloudinary', () => {
		expect(isCloudinaryUpload(original)).toBe(true);
		expect(isCloudinaryUpload('http://res.cloudinary.com/demo/image/upload/a.jpg')).toBe(false);
		expect(isCloudinaryUpload('https://res.cloudinary.com.evil.example/demo/image/upload/a.jpg')).toBe(false);
		expect(isCloudinaryUpload('https://res.cloudinary.com/demo/raw/upload/a.pdf')).toBe(false);
	});
});

describe('deliverySrcset', () => {
	it('lists each width, and is absent for images that cannot be resized', () => {
		const set = deliverySrcset(original, [400, 800]);
		expect(set).toContain('w_400,f_webp,q_100/v1699/mq/abc.jpg 400w');
		expect(set).toContain('w_800,f_webp,q_100/v1699/mq/abc.jpg 800w');
		expect(deliverySrcset('https://images.unsplash.com/x.jpg', [400])).toBeUndefined();
	});
});
