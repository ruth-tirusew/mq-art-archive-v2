// Shared foundation for comments and highlights: anchoring a reader's text selection to a
// span of an article's plain-text body, and later re-resolving that anchor against the
// article's current text. Mirrors the quote/offset strategy (and the exact re-anchoring
// algorithm) in backend/api/internal/domain/engagement/anchor.go — see that file for why
// this approach was chosen over stable block IDs.
//
// The article body renders as one <p> per `body.split('\n\n')` paragraph (see
// paragraphElements below for the exact rule, including which paragraphs get skipped),
// each tagged with a `data-paragraph-index` attribute. That index is how DOM selections
// get mapped back to an offset in the original body string, without needing a generic
// (and much trickier) DOM-Range-to-plain-text-offset walker.

export const PARAGRAPH_SEPARATOR = '\n\n';

export interface TextAnchor {
	quotedText: string;
	textOffsetStart: number;
	textOffsetEnd: number;
}

export interface AnchorResolution {
	matched: boolean;
	start: number;
	end: number;
}

/** Paragraphs actually rendered as a <p>, in the same order and with the same "skip blank
 * entries" rule the article page's template uses (`{#if paragraph.trim()}`). Each entry
 * pairs the paragraph's text with its index in the *unfiltered* split — the value that
 * belongs in a rendered element's `data-paragraph-index` attribute. */
export function paragraphElements(body: string): Array<{ index: number; text: string }> {
	return body
		.split(PARAGRAPH_SEPARATOR)
		.map((text, index) => ({ index, text }))
		.filter((p) => p.text.trim().length > 0);
}

/** The character offset each paragraph index starts at within the full body string
 * (rejoining with PARAGRAPH_SEPARATOR reproduces the original body exactly). */
export function paragraphStartOffsets(body: string): number[] {
	const parts = body.split(PARAGRAPH_SEPARATOR);
	const offsets: number[] = [];
	let running = 0;
	for (const part of parts) {
		offsets.push(running);
		running += part.length + PARAGRAPH_SEPARATOR.length;
	}
	return offsets;
}

/**
 * Convert a local offset within one paragraph (by its data-paragraph-index) into a global
 * offset into the full article body string.
 */
export function toGlobalOffset(body: string, paragraphIndex: number, localOffset: number): number {
	const offsets = paragraphStartOffsets(body);
	const base = offsets[paragraphIndex] ?? 0;
	return base + localOffset;
}

/**
 * Reads the browser's current text selection within `container` and turns it into a
 * TextAnchor with global body offsets. Returns null for an empty/collapsed selection, a
 * selection outside `container`, or one that spans multiple paragraph elements (kept out
 * of scope for this foundation — comments/highlights anchor to a single paragraph).
 *
 * Each paragraph element must carry `data-paragraph-index` matching paragraphElements()'s
 * output, and contain nothing but its own plain text (no nested inline markup yet).
 */
export function getSelectionAnchor(container: HTMLElement, body: string): TextAnchor | null {
	const selection = window.getSelection?.();
	if (!selection || selection.isCollapsed || selection.rangeCount === 0) return null;

	const range = selection.getRangeAt(0);
	const quotedText = selection.toString();
	if (!quotedText.trim()) return null;
	if (!container.contains(range.commonAncestorContainer)) return null;

	const startParagraph = closestParagraph(range.startContainer);
	const endParagraph = closestParagraph(range.endContainer);
	if (!startParagraph || !endParagraph || startParagraph !== endParagraph) return null;

	const indexAttr = startParagraph.getAttribute('data-paragraph-index');
	if (indexAttr == null) return null;
	const paragraphIndex = Number(indexAttr);

	const localStart = localTextOffset(startParagraph, range.startContainer, range.startOffset);
	const localEnd = localTextOffset(startParagraph, range.endContainer, range.endOffset);
	if (localStart == null || localEnd == null) return null;

	const start = toGlobalOffset(body, paragraphIndex, localStart);
	const end = toGlobalOffset(body, paragraphIndex, localEnd);
	return { quotedText, textOffsetStart: start, textOffsetEnd: end };
}

function closestParagraph(node: Node): HTMLElement | null {
	const el = node.nodeType === Node.ELEMENT_NODE ? (node as HTMLElement) : node.parentElement;
	return el?.closest<HTMLElement>('[data-paragraph-index]') ?? null;
}

/** Offset of (node, nodeOffset) within paragraphEl's own text content, walking its text
 * nodes in document order. paragraphEl is expected to contain a single text node today,
 * but this walks properly in case inline formatting is added later. */
function localTextOffset(paragraphEl: HTMLElement, node: Node, nodeOffset: number): number | null {
	const walker = document.createTreeWalker(paragraphEl, NodeFilter.SHOW_TEXT);
	let offset = 0;
	let current = walker.nextNode();
	while (current) {
		if (current === node) {
			return offset + nodeOffset;
		}
		offset += current.textContent?.length ?? 0;
		current = walker.nextNode();
	}
	// node may be the paragraph element itself (e.g. a boundary click) rather than a text
	// node inside it — treat offset 0 as start-of-paragraph, anything else as its end.
	if (node === paragraphEl) return nodeOffset === 0 ? 0 : paragraphEl.textContent?.length ?? null;
	return null;
}

/**
 * Re-resolve a stored anchor against the article's current body/version. Mirrors
 * Anchor.Resolve in the backend exactly: trusts the stored offsets only when the version
 * matches and they still line up with quotedText; otherwise falls back to a verbatim
 * search, and reports no match at all if the quoted text is simply gone.
 */
export function resolveAnchor(
	anchor: TextAnchor & { articleVersionAtAnchor: number },
	currentBody: string,
	currentVersion: number
): AnchorResolution {
	const { quotedText, textOffsetStart, textOffsetEnd, articleVersionAtAnchor } = anchor;

	const offsetsStillMatch =
		textOffsetStart >= 0 &&
		textOffsetStart < textOffsetEnd &&
		textOffsetEnd <= currentBody.length &&
		currentBody.slice(textOffsetStart, textOffsetEnd) === quotedText;

	if (articleVersionAtAnchor === currentVersion && offsetsStillMatch) {
		return { matched: true, start: textOffsetStart, end: textOffsetEnd };
	}

	if (!quotedText) return { matched: false, start: 0, end: 0 };

	const idx = currentBody.indexOf(quotedText);
	if (idx < 0) return { matched: false, start: 0, end: 0 };
	return { matched: true, start: idx, end: idx + quotedText.length };
}
