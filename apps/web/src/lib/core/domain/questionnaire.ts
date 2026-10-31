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
