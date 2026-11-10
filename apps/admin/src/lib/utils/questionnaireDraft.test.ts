import { describe, expect, it } from 'vitest';
import type { Question } from '$lib/core/domain/questionnaire';
import {
  blankQuestion,
  fromQuestions,
  hasProblems,
  moveItem,
  suggestKey,
  summariseChanges,
  toSaveDraftInput,
  validateDraft,
  type EditableQuestion
} from './questionnaireDraft';

const q = (over: Partial<Question> = {}): Question => ({
  key: 'bio',
  position: 1,
  type: 'long_text',
  prompt_en: 'Tell us about your practice',
  required: true,
  is_public: true,
  max_length: 5000,
  ...over
});

const editable = (over: Partial<EditableQuestion> = {}): EditableQuestion => ({
  ...fromQuestions([q()], new Set(['bio']))[0],
  ...over
});

describe('suggestKey', () => {
  it('slugifies the English prompt', () => {
    expect(suggestKey('What are your influences?', [])).toBe('what_are_your_influences');
  });
  it('keeps keys short and without a trailing underscore', () => {
    const key = suggestKey('a'.repeat(39) + ' bcd efg', []);
    expect(key.length).toBeLessThanOrEqual(40);
    expect(key.endsWith('_')).toBe(false);
  });
  it('is unique among taken keys', () => {
    expect(suggestKey('City', ['city'])).toBe('city_2');
    expect(suggestKey('City', ['city', 'city_2'])).toBe('city_3');
  });
  it('falls back for prompts with no ASCII letters (e.g. Amharic only)', () => {
    expect(suggestKey('ስለ ስራዎ', [])).toBe('question');
    expect(suggestKey('', ['question'])).toBe('question_2');
  });
});

describe('moveItem', () => {
  it('moves forward and backward without mutating', () => {
    const list = ['a', 'b', 'c'];
    expect(moveItem(list, 0, 2)).toEqual(['b', 'c', 'a']);
    expect(moveItem(list, 2, 0)).toEqual(['c', 'a', 'b']);
    expect(list).toEqual(['a', 'b', 'c']);
  });
  it('ignores out-of-range moves', () => {
    expect(moveItem(['a', 'b'], 0, 5)).toEqual(['a', 'b']);
    expect(moveItem(['a', 'b'], -1, 0)).toEqual(['a', 'b']);
  });
});

describe('fromQuestions', () => {
  it('orders by position and locks only published keys', () => {
    const list = fromQuestions([q({ key: 'b', position: 2 }), q({ key: 'a', position: 1 })], new Set(['a']));
    expect(list.map((x) => x.key)).toEqual(['a', 'b']);
    expect(list.map((x) => x.keyLocked)).toEqual([true, false]);
  });
});

describe('validateDraft', () => {
  it('accepts a good question', () => {
    expect(hasProblems(validateDraft([editable()]))).toBe(false);
  });
  it('flags a bad key, an empty prompt and a bad limit', () => {
    const bad = editable({ key: 'Bad Key', prompt_en: '  ', max_length: 0 });
    expect(validateDraft([bad])[bad.uid]).toHaveLength(3);
  });
  it('flags duplicate keys on both questions', () => {
    const a = editable({ uid: 'a' });
    const b = editable({ uid: 'b' });
    const problems = validateDraft([a, b]);
    expect(problems.a?.[0]).toMatch(/already uses/);
    expect(problems.b?.[0]).toMatch(/already uses/);
  });
  it('requires two distinct, complete options for choice questions', () => {
    const base = editable({ type: 'choice' });
    expect(validateDraft([{ ...base, options: [] }])[base.uid]).toHaveLength(1);
    const dup = { ...base, options: [{ value: 'a', label_en: 'A' }, { value: 'a', label_en: 'B' }] };
    expect(validateDraft([dup])[base.uid]?.[0]).toMatch(/used twice/);
    const blank = { ...base, options: [{ value: 'a', label_en: 'A' }, { value: '', label_en: '' }] };
    expect(validateDraft([blank])[base.uid]?.[0]).toMatch(/value and an English label/);
    const ok = { ...base, options: [{ value: 'a', label_en: 'A' }, { value: 'b', label_en: 'B' }] };
    expect(hasProblems(validateDraft([ok]))).toBe(false);
  });
});

describe('toSaveDraftInput', () => {
  it('trims text, maps blank limit to 0 and drops options on non-choice questions', () => {
    const input = toSaveDraftInput('minor', ' note ', [
      editable({ prompt_en: ' Hi ', max_length: '', options: [{ value: 'x', label_en: 'X' }] })
    ]);
    expect(input.note).toBe('note');
    expect(input.questions[0]).toMatchObject({ prompt_en: 'Hi', max_length: 0, options: [] });
  });
  it('does not send UI-only fields', () => {
    const body = JSON.stringify(toSaveDraftInput('minor', '', [editable()]));
    expect(body).not.toContain('uid');
    expect(body).not.toContain('keyLocked');
  });
});

describe('summariseChanges', () => {
  const published = [q(), q({ key: 'city', position: 2, type: 'short_text', prompt_en: 'City', required: false, max_length: 200 })];
  const keep = () => fromQuestions(published, new Set(['bio', 'city']));

  it('is unchanged and minor when nothing differs', () => {
    const s = summariseChanges(published, keep());
    expect(s.unchanged).toBe(true);
    expect(s.suggested).toBe('minor');
  });
  it('treats wording-only edits as minor', () => {
    const draft = keep();
    draft[1].prompt_en = 'Which city do you work in?';
    const s = summariseChanges(published, draft);
    expect(s.reworded).toEqual(['Which city do you work in?']);
    expect(s.suggested).toBe('minor');
    expect(s.unchanged).toBe(false);
  });
  it('treats a new question as substantive', () => {
    const draft = [...keep(), blankQuestion([])];
    draft[2].prompt_en = 'Influences';
    const s = summariseChanges(published, draft);
    expect(s.added).toEqual(['Influences']);
    expect(s.suggested).toBe('substantive');
  });
  it('treats a removed question as substantive', () => {
    const s = summariseChanges(published, keep().slice(0, 1));
    expect(s.removed).toEqual(['City']);
    expect(s.suggested).toBe('substantive');
  });
  it('treats changed rules (required, visibility, limit) as substantive', () => {
    for (const change of [{ required: true }, { is_public: false }, { max_length: 100 as const }]) {
      const draft = keep();
      Object.assign(draft[1], change);
      expect(summariseChanges(published, draft).suggested).toBe('substantive');
    }
  });
  it('notices a pure reorder', () => {
    const s = summariseChanges(published, moveItem(keep(), 0, 1));
    expect(s.unchanged).toBe(false);
    expect(s.suggested).toBe('minor');
  });
  it('counts a blank limit as "keep the current limit", not a change', () => {
    const draft = keep();
    draft[0].max_length = '';
    expect(summariseChanges(published, draft).unchanged).toBe(true);
  });
});
