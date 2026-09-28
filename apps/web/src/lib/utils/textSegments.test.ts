import { describe, expect, it } from 'vitest';
import { segmentParagraph } from './textSegments';

describe('segmentParagraph', () => {
	it('returns the whole paragraph untagged when no spans apply', () => {
		expect(segmentParagraph('hello world', 0, {})).toEqual([{ text: 'hello world', tags: [] }]);
	});

	it('tags a single span within the paragraph', () => {
		// global offsets 6-11 = "world" when the paragraph starts at global offset 0
		const result = segmentParagraph('hello world', 0, { mine: [{ start: 6, end: 11 }] });
		expect(result).toEqual([
			{ text: 'hello ', tags: [] },
			{ text: 'world', tags: ['mine'] }
		]);
	});

	it('clips a span that starts before or ends after the paragraph', () => {
		// paragraph occupies global 10-21; span 5-15 should clip to local 0-5
		const result = segmentParagraph('hello world', 10, { mine: [{ start: 5, end: 15 }] });
		expect(result).toEqual([
			{ text: 'hello', tags: ['mine'] },
			{ text: ' world', tags: [] }
		]);
	});

	it('combines overlapping tags from different categories on the shared run', () => {
		const result = segmentParagraph('hello world', 0, {
			mine: [{ start: 0, end: 8 }],
			popular: [{ start: 6, end: 11 }]
		});
		expect(result).toEqual([
			{ text: 'hello ', tags: ['mine'] },
			{ text: 'wo', tags: ['mine', 'popular'] },
			{ text: 'rld', tags: ['popular'] }
		]);
	});

	it('ignores spans entirely outside the paragraph', () => {
		const result = segmentParagraph('hello world', 100, { mine: [{ start: 0, end: 5 }] });
		expect(result).toEqual([{ text: 'hello world', tags: [] }]);
	});
});
