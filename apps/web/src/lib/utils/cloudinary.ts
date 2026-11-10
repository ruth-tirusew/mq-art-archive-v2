/**
 * Delivery URLs for images stored in Cloudinary. The original file is kept exactly as uploaded;
 * these only change what is sent to the browser: WebP at quality 100, which Cloudinary treats as
 * lossless, at (at most) the width asked for. Images that don't live in Cloudinary (seed data,
 * other CDNs) are returned untouched.
 */

const UPLOAD_MARKER = '/image/upload/';
const LOSSLESS = 'f_webp,q_100';

export function isCloudinaryUpload(url: string): boolean {
	try {
		const u = new URL(url);
		return u.protocol === 'https:' && u.hostname === 'res.cloudinary.com' && u.pathname.includes(UPLOAD_MARKER);
	} catch {
		return false;
	}
}

/** The URL to show an image at, at most `width` CSS pixels wide (pass 2x for sharp screens). */
export function deliveryUrl(url: string, width?: number): string {
	if (!isCloudinaryUpload(url)) return url;
	const u = new URL(url);
	const at = u.pathname.indexOf(UPLOAD_MARKER) + UPLOAD_MARKER.length;
	// Already transformed (someone stored a delivery URL)? Leave it alone rather than stacking.
	if (/^[a-z]{1,3}_[^/]+\//.test(u.pathname.slice(at)) && !u.pathname.slice(at).startsWith('v')) return url;
	const transformation = width && width > 0 ? `c_limit,w_${Math.round(width)},${LOSSLESS}` : LOSSLESS;
	u.pathname = `${u.pathname.slice(0, at)}${transformation}/${u.pathname.slice(at)}`;
	return u.toString();
}

/** A srcset of delivery URLs, so the browser picks the smallest width that is still sharp. */
export function deliverySrcset(url: string, widths: number[]): string | undefined {
	if (!isCloudinaryUpload(url)) return undefined;
	return widths.map((w) => `${deliveryUrl(url, w)} ${w}w`).join(', ');
}
