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
