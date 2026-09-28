import { describe, expect, it } from 'vitest';
import {
	paragraphElements,
	paragraphStartOffsets,
	toGlobalOffset,
	resolveAnchor
} from './textAnchor';

describe('paragraphElements', () => {
	it('skips blank paragraphs but keeps their original split index', () => {
		const body = 'First.\n\n\n\nSecond.';
		expect(paragraphElements(body)).toEqual([
			{ index: 0, text: 'First.' },
			{ index: 2, text: 'Second.' }
		]);
	});
});

describe('paragraphStartOffsets', () => {
	it('accounts for the \\n\\n separator between paragraphs', () => {
		const body = 'One.\n\nTwo.\n\nThree.';
		expect(paragraphStartOffsets(body)).toEqual([0, 6, 12]);
		expect(body.slice(6, 9)).toBe('Two');
		expect(body.slice(12, 17)).toBe('Three');
	});
});

describe('toGlobalOffset', () => {
	it('maps a local paragraph offset to the full-body offset', () => {
		const body = 'One.\n\nTwo.\n\nThree.';
		expect(toGlobalOffset(body, 2, 0)).toBe(12);
		expect(toGlobalOffset(body, 0, 2)).toBe(2);
	});
});

describe('resolveAnchor', () => {
	const anchor = (over: Partial<Parameters<typeof resolveAnchor>[0]> = {}) => ({
		quotedText: 'brown fox',
		textOffsetStart: 10,
		textOffsetEnd: 19,
		articleVersionAtAnchor: 3,
		...over
	});

	it('trusts stored offsets when the version is unchanged', () => {
		const body = 'The quick brown fox jumps over the lazy dog.';
		expect(resolveAnchor(anchor(), body, 3)).toEqual({ matched: true, start: 10, end: 19 });
	});

	it('falls back to a verbatim search after an edit shifts offsets', () => {
		const original = 'The quick brown fox jumps over the lazy dog.';
		const edited = 'An intro paragraph.\n\n' + original;
		const want = 'An intro paragraph.\n\n'.length + 10;
		expect(resolveAnchor(anchor(), edited, 4)).toEqual({
			matched: true,
			start: want,
			end: want + 'brown fox'.length
		});
	});

	it('reports no match once the quoted text is gone', () => {
		expect(resolveAnchor(anchor({ quotedText: 'this text was removed' }), 'different content', 2)).toEqual(
			{ matched: false, start: 0, end: 0 }
		);
	});

	it('never matches an empty quoted text', () => {
		expect(resolveAnchor(anchor({ quotedText: '' }), 'anything', 2)).toEqual({
			matched: false,
			start: 0,
			end: 0
		});
	});
});
