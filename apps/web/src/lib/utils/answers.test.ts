import { describe, expect, it } from 'vitest';
import type { PublicAnswer } from '$lib/core/domain/questionnaire';
import { answerText, linkLabel, safeHttpUrl } from './answers';

describe('safeHttpUrl', () => {
	it('accepts and normalises http(s) links', () => {
		expect(safeHttpUrl('https://example.com')).toBe('https://example.com/');
		expect(safeHttpUrl('  http://example.com/a?b=c#d  ')).toBe('http://example.com/a?b=c#d');
	});

	it('returns what the browser would actually navigate to', () => {
		// WHATWG parsing ignores the extra slash. The result is still a plain http(s) link; the API's
		// stricter validator is what stops a value like this from being stored in the first place.
		expect(safeHttpUrl('https:///nohost')).toBe('https://nohost/');
	});

	it.each([
		'javascript:alert(1)',
		'JaVaScRiPt:alert(1)',
		' javascript:alert(1)',
		'java\tscript:alert(1)',
		'data:text/html,<script>alert(1)</script>',
		'vbscript:msgbox(1)',
		'file:///etc/passwd',
		'//evil.example/x',
		'example.com',
		'/relative/path',
		'https://user:pw@example.com',
		'mailto:a@b.co',
		'',
		'   '
	])('rejects %j', (value) => {
		expect(safeHttpUrl(value)).toBeNull();
	});
});

describe('linkLabel', () => {
	it('shows host and path without scheme or trailing slash', () => {
		expect(linkLabel('https://example.com/')).toBe('example.com');
		expect(linkLabel('https://example.com/work/')).toBe('example.com/work');
	});
});

describe('answerText', () => {
	const choice: PublicAnswer = {
		key: 'medium', type: 'choice', prompt_en: 'Medium', body: 'ink', format: 'plain', lang: 'en',
		options: [{ value: 'oil', label_en: 'Oil' }, { value: 'ink', label_en: 'Ink' }]
	};
	it('shows the option label for a choice', () => {
		expect(answerText(choice)).toBe('Ink');
	});
	it('falls back to the stored value when the option is gone', () => {
		expect(answerText({ ...choice, body: 'gone' })).toBe('gone');
	});
	it('returns the body for other types', () => {
		expect(answerText({ ...choice, type: 'short_text', body: 'hello' })).toBe('hello');
	});
});

import type { FormQuestion } from '$lib/core/domain/questionnaire';
import {
	answerLength,
	answerProblem,
	detectLang,
	editorValue,
	formatFor,
	plainToHtml,
	truncateText,
	htmlText,
	isBlankAnswer,
	missingRequired
} from './answers';

const fq = (over: Partial<FormQuestion> = {}): FormQuestion => ({
	key: 'k', position: 1, type: 'short_text', prompt_en: 'Q', required: false, is_public: true, max_length: 10, ...over
});

describe('detectLang', () => {
	it('classifies by letters', () => {
		expect(detectLang('Hello there')).toBe('en');
		expect(detectLang('ሰላም ዓለም')).toBe('am');
		expect(detectLang('I paint ሰላም and more words here')).toBe('mixed');
	});
	it('ignores digits, punctuation and spaces', () => {
		expect(detectLang('ሰላም, 2026!')).toBe('am');
		expect(detectLang('')).toBe('en');
		expect(detectLang('1234 !!!')).toBe('en');
	});
});

describe('htmlText / isBlankAnswer / answerLength', () => {
	it('extracts visible text', () => {
		expect(htmlText('<p>Tom &amp; <strong>Jerry</strong></p><p>Second</p>')).toBe('Tom & Jerry Second');
		expect(htmlText('<ul><li>a</li><li>b</li></ul>')).toBe('a b');
	});
	it('treats empty markup as blank', () => {
		expect(isBlankAnswer('<p></p>', 'rich')).toBe(true);
		expect(isBlankAnswer('<p><br></p>', 'rich')).toBe(true);
		expect(isBlankAnswer('<p>x</p>', 'rich')).toBe(false);
		expect(isBlankAnswer('  ', 'plain')).toBe(true);
	});
	it('counts characters, not bytes or tags', () => {
		expect(answerLength('<p><strong>ሰላም</strong></p>', 'rich')).toBe(3);
		expect(answerLength('ሰላም', 'plain')).toBe(3);
	});
});

describe('answerProblem', () => {
	it('allows blank (that clears an answer)', () => {
		expect(answerProblem(fq(), '')).toBeNull();
		expect(answerProblem(fq({ type: 'long_text' }), '<p></p>')).toBeNull();
	});
	it('flags over-long answers, counting visible characters', () => {
		expect(answerProblem(fq(), 'a'.repeat(11))).toMatch(/10-character limit/);
		expect(answerProblem(fq(), 'a'.repeat(10))).toBeNull();
		expect(answerProblem(fq({ type: 'long_text' }), `<p><strong>${'a'.repeat(11)}</strong></p>`)).toMatch(/limit/);
	});
	it('requires http(s) links', () => {
		const link = fq({ type: 'link', max_length: 300 });
		expect(answerProblem(link, 'javascript:alert(1)')).toMatch(/http/);
		expect(answerProblem(link, 'example.com')).toMatch(/http/);
		expect(answerProblem(link, 'https://example.com')).toBeNull();
	});
	it('requires a listed option for choices', () => {
		const choice = fq({ type: 'choice', options: [{ value: 'a', label_en: 'A' }] });
		expect(answerProblem(choice, 'zzz')).toMatch(/Choose/);
		expect(answerProblem(choice, 'a')).toBeNull();
	});
});

describe('formatFor / missingRequired', () => {
	it('stores only long answers as rich text', () => {
		expect(formatFor(fq({ type: 'long_text' }))).toBe('rich');
		expect(formatFor(fq({ type: 'short_text' }))).toBe('rich');
		expect(formatFor(fq({ type: 'link' }))).toBe('plain');
		expect(formatFor(fq({ type: 'choice' }))).toBe('plain');
	});
	it('lists blank required questions in question order', () => {
		const qs = [
			fq({ key: 'b', position: 2, required: true }),
			fq({ key: 'a', position: 1, required: true, type: 'long_text' }),
			fq({ key: 'c', position: 3, required: false })
		];
		expect(missingRequired(qs, { a: '<p></p>', b: 'x' })).toEqual(['a']);
		expect(missingRequired(qs, {})).toEqual(['a', 'b']);
		expect(missingRequired(qs, { a: '<p>hi</p>', b: 'x' })).toEqual([]);
	});
});

describe('plainToHtml / editorValue', () => {
	it('escapes plain text and keeps its paragraphs and line breaks', () => {
		expect(plainToHtml('a < b & c\nnext line\n\nsecond')).toBe('<p>a &lt; b &amp; c<br>next line</p><p>second</p>');
		expect(plainToHtml('   ')).toBe('');
	});

	it('converts an old plain answer for a question that is now rich text', () => {
		expect(editorValue(fq({ type: 'short_text' }), { body: 'Oil <3', format: 'plain' })).toBe('<p>Oil &lt;3</p>');
	});

	it('leaves rich answers, links and choices as they are', () => {
		expect(editorValue(fq({ type: 'short_text' }), { body: '<p><em>Oil</em></p>', format: 'rich' })).toBe('<p><em>Oil</em></p>');
		expect(editorValue(fq({ type: 'link' }), { body: 'https://x.et', format: 'plain' })).toBe('https://x.et');
		expect(editorValue(fq({ type: 'short_text' }), undefined)).toBe('');
	});
});

describe('truncateText', () => {
	it('leaves short text alone and collapses stray whitespace', () => {
		expect(truncateText('  A  short\n bio. ', 140)).toBe('A short bio.');
	});

	it('cuts at a word boundary and marks the cut', () => {
		const out = truncateText('I paint light and memory in the highlands of Ethiopia every morning', 30);
		expect(out).toBe('I paint light and memory in…');
		expect(Array.from(out).length).toBeLessThanOrEqual(31);
	});

	it('never leaves dangling punctuation before the ellipsis', () => {
		expect(truncateText('Painter, sculptor, and teacher, based in Addis', 24)).toBe('Painter, sculptor, and…');
	});

	it('counts characters, not bytes, so Amharic is not cut early', () => {
		const am = 'ሰላም '.repeat(10).trim(); // 39 characters, 3 bytes each
		expect(truncateText(am, 40)).toBe(am);
		expect(truncateText(am, 20).endsWith('…')).toBe(true);
	});
});

