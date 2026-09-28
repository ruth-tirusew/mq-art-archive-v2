// Shared rendering primitive for highlights and comments: splitting one paragraph's text
// into runs tagged with which spans (by category — "mine", "popular", "comment", ...)
// cover them, so a template can apply combined CSS classes per run instead of one
// exclusive style per span. Spans are given in GLOBAL body-string offsets (the same
// space $lib/utils/textAnchor computes) and clipped to the paragraph automatically.

export interface Span {
	start: number;
	end: number;
}

export interface TextSegment {
	text: string;
	tags: string[];
}

/**
 * Splits paragraphText (which occupies [paragraphGlobalStart, paragraphGlobalStart +
 * paragraphText.length) in the full body) into ordered, non-overlapping runs, each tagged
 * with every category from spansByTag whose span covers that run.
 */
export function segmentParagraph(
	paragraphText: string,
	paragraphGlobalStart: number,
	spansByTag: Record<string, Span[]>
): TextSegment[] {
	const paragraphEnd = paragraphGlobalStart + paragraphText.length;
	const boundaries = new Set<number>([0, paragraphText.length]);
	const localSpans: Array<{ start: number; end: number; tag: string }> = [];

	for (const [tag, spans] of Object.entries(spansByTag)) {
		for (const span of spans) {
			const start = Math.max(span.start, paragraphGlobalStart) - paragraphGlobalStart;
			const end = Math.min(span.end, paragraphEnd) - paragraphGlobalStart;
			if (start < end) {
				localSpans.push({ start, end, tag });
				boundaries.add(start);
				boundaries.add(end);
			}
		}
	}

	const sorted = [...boundaries].sort((a, b) => a - b);
	const segments: TextSegment[] = [];
	for (let i = 0; i < sorted.length - 1; i++) {
		const start = sorted[i];
		const end = sorted[i + 1];
		if (start >= end) continue;
		const tags = localSpans.filter((s) => s.start <= start && s.end >= end).map((s) => s.tag);
		segments.push({ text: paragraphText.slice(start, end), tags });
	}
	return segments;
}
