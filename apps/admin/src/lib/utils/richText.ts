/**
 * Small helpers for the rich-text editor, the same rules as the web app's (apps/web/src/lib/utils/answers.ts).
 * The server sanitises everything again on save; these only keep the editor honest.
 */

/**
 * Returns a normalised URL only for plain http(s) links with a host and no embedded credentials,
 * otherwise null: javascript:, data:, protocol-relative and bare-host values all fail.
 */
export function safeHttpUrl(value: string): string | null {
  const raw = value.trim();
  if (!raw) return null;
  let url: URL;
  try {
    url = new URL(raw);
  } catch {
    return null;
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') return null;
  if (!url.hostname || url.username || url.password) return null;
  return url.href;
}

const escapeHtml = (s: string) =>
  s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

/** Plain text as editor HTML: blank-line-separated paragraphs, single newlines as <br>, all escaped. */
export function plainToHtml(text: string): string {
  return text
    .split(/\n{2,}/)
    .map((p) => p.trim())
    .filter(Boolean)
    .map((p) => `<p>${escapeHtml(p).replace(/\n/g, '<br>')}</p>`)
    .join('');
}

/** The visible words of an editor's HTML, for previews and "is this blank" checks. Never for display as HTML. */
export function htmlText(html: string): string {
  return html
    .replace(/<br\s*\/?>/gi, ' ')
    .replace(/<\/(p|li|blockquote|h2|h3)>/gi, ' ')
    .replace(/<[^>]*>/g, '')
    .replace(/&nbsp;/g, ' ')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&amp;/g, '&')
    .replace(/\s+/g, ' ')
    .trim();
}
