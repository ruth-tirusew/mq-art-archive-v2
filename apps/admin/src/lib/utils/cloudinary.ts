/**
 * Resized, lossless delivery URL for an image stored in Cloudinary (WebP at quality 100); anything
 * else is returned as it is. The same rule as apps/web/src/lib/utils/cloudinary.ts, kept small and
 * duplicated until shared front-end code moves into packages/.
 */
const MARKER = '/image/upload/';

export function deliveryUrl(url: string, width: number): string {
  try {
    const u = new URL(url);
    if (u.protocol !== 'https:' || u.hostname !== 'res.cloudinary.com' || !u.pathname.includes(MARKER)) return url;
    const at = u.pathname.indexOf(MARKER) + MARKER.length;
    u.pathname = `${u.pathname.slice(0, at)}c_limit,w_${Math.round(width)},f_webp,q_100/${u.pathname.slice(at)}`;
    return u.toString();
  } catch {
    return url;
  }
}
