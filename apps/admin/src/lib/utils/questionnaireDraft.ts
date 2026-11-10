import type {
  ChangeKind,
  Question,
  QuestionInput,
  QuestionOption,
  QuestionType,
  SaveDraftInput
} from '$lib/core/domain/questionnaire';

/** A question while it is being edited. `uid` is only for keying the UI list; it is never sent. */
export interface EditableQuestion {
  uid: string;
  key: string;
  /** True for keys already published: renaming would orphan every artist's existing answer. */
  keyLocked: boolean;
  type: QuestionType;
  prompt_en: string;
  prompt_am: string;
  help_en: string;
  help_am: string;
  required: boolean;
  is_public: boolean;
  /** Empty means "use the default for this type". */
  max_length: number | '';
  options: QuestionOption[];
}

export const KEY_PATTERN = /^[a-z0-9_]{1,64}$/;

let uidCounter = 0;
export function newUid(): string {
  uidCounter += 1;
  return `q${uidCounter}`;
}

export function blankQuestion(taken: Iterable<string>): EditableQuestion {
  return {
    uid: newUid(),
    key: suggestKey('', taken),
    keyLocked: false,
    type: 'short_text',
    prompt_en: '',
    prompt_am: '',
    help_en: '',
    help_am: '',
    required: false,
    is_public: true,
    max_length: '',
    options: []
  };
}

export function fromQuestions(questions: Question[], publishedKeys: ReadonlySet<string>): EditableQuestion[] {
  return [...questions]
    .sort((a, b) => a.position - b.position)
    .map((q) => ({
      uid: newUid(),
      key: q.key,
      keyLocked: publishedKeys.has(q.key),
      type: q.type,
      prompt_en: q.prompt_en,
      prompt_am: q.prompt_am ?? '',
      help_en: q.help_en ?? '',
      help_am: q.help_am ?? '',
      required: q.required,
      is_public: q.is_public,
      max_length: q.max_length,
      options: (q.options ?? []).map((o) => ({ value: o.value, label_en: o.label_en, label_am: o.label_am ?? '' }))
    }));
}

/**
 * Derive a key from the English prompt: lowercase ASCII words joined by underscores, at most 40
 * characters, and unique among `taken`. Falls back to "question" for prompts with no ASCII letters.
 */
export function suggestKey(prompt: string, taken: Iterable<string>): string {
  const used = new Set(taken);
  const base =
    prompt
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '')
      .slice(0, 40)
      .replace(/_+$/, '') || 'question';
  if (!used.has(base)) return base;
  for (let n = 2; ; n++) {
    const candidate = `${base}_${n}`;
    if (!used.has(candidate)) return candidate;
  }
}

/** Returns a new array with the item at `from` moved to `to`. Out-of-range moves return the input unchanged. */
export function moveItem<T>(list: readonly T[], from: number, to: number): T[] {
  if (from === to || from < 0 || to < 0 || from >= list.length || to >= list.length) return [...list];
  const copy = [...list];
  const [item] = copy.splice(from, 1);
  copy.splice(to, 0, item);
  return copy;
}

export type DraftProblems = Record<string, string[]>;

/** Mirrors the server's rules so most mistakes are caught before a round trip. The server stays the authority. */
export function validateDraft(questions: readonly EditableQuestion[]): DraftProblems {
  const problems: DraftProblems = {};
  const add = (uid: string, message: string) => (problems[uid] ??= []).push(message);

  const seen = new Map<string, string>();
  for (const q of questions) {
    if (!KEY_PATTERN.test(q.key)) add(q.uid, 'The key must be 1–64 lowercase letters, numbers or underscores.');
    const first = seen.get(q.key);
    if (first !== undefined) {
      add(q.uid, `Another question already uses the key "${q.key}".`);
      add(first, `Another question already uses the key "${q.key}".`);
    } else {
      seen.set(q.key, q.uid);
    }
    if (!q.prompt_en.trim()) add(q.uid, 'The English question text is required.');
    if (q.max_length !== '' && (!Number.isInteger(q.max_length) || q.max_length < 1 || q.max_length > 20000)) {
      add(q.uid, 'The length limit must be a whole number between 1 and 20,000.');
    }
    if (q.type === 'choice') {
      const values = new Set<string>();
      if (q.options.length < 2) add(q.uid, 'A multiple-choice question needs at least two options.');
      for (const o of q.options) {
        if (!o.value.trim() || !o.label_en.trim()) {
          add(q.uid, 'Every option needs a value and an English label.');
          break;
        }
      }
      for (const o of q.options) {
        if (values.has(o.value)) {
          add(q.uid, `The option value "${o.value}" is used twice.`);
          break;
        }
        values.add(o.value);
      }
    }
  }
  return problems;
}

export function hasProblems(problems: DraftProblems): boolean {
  return Object.keys(problems).length > 0;
}

export function toSaveDraftInput(
  changeKind: ChangeKind,
  note: string,
  questions: readonly EditableQuestion[]
): SaveDraftInput {
  return {
    change_kind: changeKind,
    note: note.trim(),
    questions: questions.map(
      (q): QuestionInput => ({
        key: q.key.trim(),
        type: q.type,
        prompt_en: q.prompt_en.trim(),
        prompt_am: q.prompt_am.trim(),
        help_en: q.help_en.trim(),
        help_am: q.help_am.trim(),
        required: q.required,
        is_public: q.is_public,
        max_length: q.max_length === '' ? 0 : q.max_length,
        options: q.type === 'choice' ? q.options.map((o) => ({ ...o, label_am: o.label_am?.trim() ?? '' })) : []
      })
    )
  };
}

export interface ChangeSummary {
  added: string[];
  removed: string[];
  /** Same key, different wording or help text only. */
  reworded: string[];
  /** Same key, but required / visibility / options / limit changed. */
  rules: string[];
  suggested: ChangeKind;
  /** True when nothing differs from the published revision. */
  unchanged: boolean;
}

const norm = (s: string | undefined) => (s ?? '').trim();

/**
 * Compare the draft with the latest published questions, to describe what a publish would change
 * and recommend minor (wording only) or substantive (a question added or removed, or its rules
 * changed), which decides whether artists can be nudged.
 */
export function summariseChanges(published: readonly Question[], draft: readonly EditableQuestion[]): ChangeSummary {
  const before = new Map(published.map((q) => [q.key, q]));
  const after = new Map(draft.map((q) => [q.key, q]));
  const label = (key: string, prompt: string) => `${prompt.trim() || key}`;

  const added = draft.filter((q) => !before.has(q.key)).map((q) => label(q.key, q.prompt_en));
  const removed = published.filter((q) => !after.has(q.key)).map((q) => label(q.key, q.prompt_en));
  const reworded: string[] = [];
  const rules: string[] = [];

  for (const q of draft) {
    const old = before.get(q.key);
    if (!old) continue;
    const oldLimit = old.max_length;
    const newLimit = q.max_length === '' ? oldLimit : q.max_length;
    const optionsChanged =
      JSON.stringify((old.options ?? []).map((o) => [o.value, norm(o.label_en), norm(o.label_am)])) !==
      JSON.stringify(q.options.map((o) => [o.value, norm(o.label_en), norm(o.label_am)]));
    if (
      old.type !== q.type ||
      old.required !== q.required ||
      old.is_public !== q.is_public ||
      oldLimit !== newLimit ||
      (q.type === 'choice' && optionsChanged)
    ) {
      rules.push(label(q.key, q.prompt_en));
    } else if (
      norm(old.prompt_en) !== norm(q.prompt_en) ||
      norm(old.prompt_am) !== norm(q.prompt_am) ||
      norm(old.help_en) !== norm(q.help_en) ||
      norm(old.help_am) !== norm(q.help_am)
    ) {
      reworded.push(label(q.key, q.prompt_en));
    }
  }

  const orderChanged =
    published.length === draft.length &&
    [...published].sort((a, b) => a.position - b.position).some((q, i) => draft[i]?.key !== q.key);

  const substantive = added.length > 0 || removed.length > 0 || rules.length > 0;
  return {
    added,
    removed,
    reworded,
    rules,
    suggested: substantive ? 'substantive' : 'minor',
    unchanged: !substantive && reworded.length === 0 && !orderChanged
  };
}
