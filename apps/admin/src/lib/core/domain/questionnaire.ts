export type QuestionType = 'short_text' | 'long_text' | 'choice' | 'link';
export type ChangeKind = 'minor' | 'substantive';
export type AnswerFormat = 'plain' | 'rich';
export type AnswerLang = 'en' | 'am' | 'mixed';

export const questionTypeLabels: Record<QuestionType, string> = {
  short_text: 'Short answer',
  long_text: 'Long answer (formatted)',
  choice: 'Multiple choice',
  link: 'Link'
};

export interface QuestionOption {
  value: string;
  label_en: string;
  label_am?: string;
}

export interface Question {
  key: string;
  position: number;
  type: QuestionType;
  prompt_en: string;
  prompt_am?: string;
  help_en?: string;
  help_am?: string;
  required: boolean;
  is_public: boolean;
  max_length: number;
  options?: QuestionOption[];
}

export interface QuestionnaireRevision {
  id: string;
  number?: number;
  status: 'draft' | 'published';
  change_kind: ChangeKind;
  note?: string;
  published_at?: string;
  created_at: string;
  updated_at: string;
  questions: Question[];
}

/** A question as sent to the API: position is implied by order, max_length 0 means the type's default. */
export interface QuestionInput {
  key: string;
  type: QuestionType;
  prompt_en: string;
  prompt_am: string;
  help_en: string;
  help_am: string;
  required: boolean;
  is_public: boolean;
  max_length: number;
  options: QuestionOption[];
}

export interface SaveDraftInput {
  change_kind: ChangeKind;
  note: string;
  questions: QuestionInput[];
  /** The draft's updated_at when the editor loaded it; a save made against an older copy is refused (409). */
  expected_updated_at?: string;
}

export interface Answer {
  key: string;
  body: string;
  format: AnswerFormat;
  lang: AnswerLang;
  updated_at: string;
}

export interface ReviewItem {
  question: Question;
  answer: Answer | null;
}

export interface ReviewView {
  revision: QuestionnaireRevision;
  items: ReviewItem[];
  /** Answers to questions that are not in this revision (added later, or since removed). */
  other: Answer[];
}

export interface NudgePreview {
  revision_id: string;
  revision_number?: number;
  change_kind: ChangeKind;
  eligible: number;
  can_send: boolean;
}
