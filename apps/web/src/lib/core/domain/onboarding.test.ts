import { describe, expect, it } from 'vitest';
import { ApiError } from '$lib/adapters/api/client';
import { errorKind, kindForStatus } from '$lib/core/domain/errors';
import { composeLinkNotes, parseLinkNotes, isValidHandle, normalizeHandle } from '$lib/core/domain/onboarding';

describe('handles', () => {
	it('normalizes case and whitespace before validating', () => {
		expect(normalizeHandle('  Ada_1 ')).toBe('ada_1');
		expect(isValidHandle(normalizeHandle(' Ada_1 '))).toBe(true);
	});
	it('enforces 3–30 lowercase letters, digits and underscores', () => {
		for (const bad of ['ab', 'a'.repeat(31), 'has space', 'dash-ed', 'UPPER', '']) expect(isValidHandle(bad)).toBe(false);
		for (const ok of ['abc', 'a'.repeat(30), 'a_1']) expect(isValidHandle(ok)).toBe(true);
	});
});

describe('composeLinkNotes', () => {
	it('keeps only the filled fields, labelled, in a stable order', () => {
		expect(composeLinkNotes({ instagram: ' @ada ', website: '', telegram: '', about: ' I paint ' })).toBe(
			'Instagram: @ada\nAbout my practice:\nI paint'
		);
		expect(composeLinkNotes({ instagram: '', website: '', telegram: '', about: '' })).toBe('');
	});
});

describe('error kinds', () => {
	it('maps HTTP statuses to kinds', () => {
		expect(kindForStatus(404)).toBe('not_found');
		expect(kindForStatus(409)).toBe('conflict');
		expect(kindForStatus(403)).toBe('forbidden');
		expect(kindForStatus(429)).toBe('rate_limited');
		expect(kindForStatus(400)).toBe('invalid');
		expect(kindForStatus(500)).toBe('unknown');
	});
	it('lets screens branch on kind without knowing about HTTP', () => {
		expect(errorKind(new ApiError('nope', 409))).toBe('conflict');
		expect(errorKind(new Error('plain'))).toBe('unknown');
		expect(errorKind('weird')).toBe('unknown');
	});
});

describe('parseLinkNotes', () => {
	const empty = { instagram: '', website: '', telegram: '', about: '' };

	it('round-trips what composeLinkNotes produced, so re-applying never nests the labels', () => {
		const fields = { instagram: '@ada', website: 'https://ada.art', telegram: '@ada_t', about: 'I paint.\nSince 2010.' };
		expect(parseLinkNotes(composeLinkNotes(fields))).toEqual(fields);
		const again = composeLinkNotes(parseLinkNotes(composeLinkNotes(fields)));
		expect(again).toBe(composeLinkNotes(fields));
	});

	it('handles notes with only links, or only an about', () => {
		expect(parseLinkNotes('Instagram: @ada')).toEqual({ ...empty, instagram: '@ada' });
		expect(parseLinkNotes('About my practice:\nhello')).toEqual({ ...empty, about: 'hello' });
		expect(parseLinkNotes('')).toEqual(empty);
	});

	it('keeps free text it does not recognise as the about text', () => {
		expect(parseLinkNotes('Just a plain note')).toEqual({ ...empty, about: 'Just a plain note' });
	});
});
