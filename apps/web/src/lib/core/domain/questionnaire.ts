export type QuestionType = 'short_text' | 'long_text' | 'choice' | 'link';
export type AnswerFormat = 'plain' | 'rich';
export type AnswerLang = 'en' | 'am' | 'mixed';

export interface QuestionOption {
	value: string;
	label_en: string;
	label_am?: string;
}

/**
 * One visible answer on an artist's public page. For format "rich" the body is HTML the API has
 * already sanitised; for "plain" it is text and must be rendered as text. A choice answer's body is
 * the option's value, and `options` lets the page show the label.
 */
export interface PublicAnswer {
	key: string;
	type: QuestionType;
	prompt_en: string;
	prompt_am?: string;
	options?: QuestionOption[];
	body: string;
	format: AnswerFormat;
	lang: AnswerLang;
}

/** A question on the form an applicant or artist fills in. */
export interface FormQuestion {
	key: string;
	position: number;
	type: QuestionType;
	prompt_en: string;
	prompt_am?: string;
	help_en?: string;
	help_am?: string;
	required: boolean;
	/** False means only the review team sees the answer, never the public profile. */
	is_public: boolean;
	max_length: number;
	options?: QuestionOption[];
}

export interface FormRevision {
	id: string;
	number?: number;
	questions: FormQuestion[];
}

export interface SavedAnswer {
	key: string;
	body: string;
	format: AnswerFormat;
	lang: AnswerLang;
	updated_at: string;
}

export interface QuestionnaireForm {
	revision: FormRevision;
	answers: SavedAnswer[];
	/** True while an application is under review; answers cannot change. */
	locked: boolean;
}

/** What the client sends when saving one answer. */
export interface AnswerInput {
	key: string;
	body: string;
	format: AnswerFormat;
	lang: AnswerLang;
}

/** An optional invitation to answer questions added since the artist last answered. */
export interface QuestionnaireNudge {
	revision_id: string;
	sent_at: string;
	questions: FormQuestion[];
}
