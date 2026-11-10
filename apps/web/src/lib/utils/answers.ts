import type {
	AnswerFormat,
	AnswerLang,
	FormQuestion,
	PublicAnswer
} from '$lib/core/domain/questionnaire';

/**
 * Returns a normalised URL only for plain http(s) links with a host and no embedded credentials,
 * otherwise null. Link answers come from users, so this is the gate before anything becomes an
 * href: javascript:, data:, vbscript:, file:, protocol-relative and bare-host values all fail.
 */
export function safeHttpUrl(value: string): string | null {
	const raw = value.trim();
	if (!raw) return null;
	let url: URL;
	try {
		url = new URL(raw);
	} catch {
		return null;
	}
	if (url.protocol !== 'http:' && url.protocol !== 'https:') return null;
	if (!url.hostname || url.username || url.password) return null;
	return url.href;
}

/** A short, readable form of a URL for link text: host and path without scheme or trailing slash. */
export function linkLabel(href: string): string {
	try {
		const url = new URL(href);
		return `${url.host}${url.pathname === '/' ? '' : url.pathname}`.replace(/\/$/, '');
	} catch {
		return href;
	}
}

/** The text of a plain or choice answer. A choice is stored as the option's value; show its label. */
export function answerText(answer: PublicAnswer): string {
	if (answer.type === 'choice') {
		return answer.options?.find((o) => o.value === answer.body)?.label_en ?? answer.body;
	}
	return answer.body;
}


const ETHIOPIC = /[ሀ-፿ᎀ-᎟ⶀ-⷟꬀-꬯]/gu;
const LETTER = /\p{L}/gu;

/**
 * Which language an answer is written in, judged from its letters: all Ethiopic is Amharic, a mix
 * is "mixed", none is English. Artists never choose; it only drives the `lang` attribute (font and
 * screen-reader pronunciation). Digits, punctuation and spaces are ignored.
 */
export function detectLang(text: string): AnswerLang {
	const letters = text.match(LETTER)?.length ?? 0;
	if (letters === 0) return 'en';
	const ethiopic = text.match(ETHIOPIC)?.length ?? 0;
	if (ethiopic === 0) return 'en';
	return ethiopic / letters >= 0.9 ? 'am' : 'mixed';
}

/** The visible text of an editor's HTML. Used only for emptiness and length, never for display. */
export function htmlText(html: string): string {
	return html
		.replace(/<br\s*\/?>/gi, ' ')
		.replace(/<\/(p|li|blockquote|h2|h3)>/gi, ' ')
		.replace(/<[^>]*>/g, '')
		.replace(/&nbsp;/g, ' ')
		.replace(/&lt;/g, '<')
		.replace(/&gt;/g, '>')
		.replace(/&quot;/g, '"')
		.replace(/&#39;/g, "'")
		.replace(/&amp;/g, '&')
		.replace(/\s+/g, ' ')
		.trim();
}

export function isBlankAnswer(body: string, format: AnswerFormat): boolean {
	return (format === 'rich' ? htmlText(body) : body).trim() === '';
}

/** Visible characters (not bytes, not markup), which is how the server counts too. */
export function answerLength(body: string, format: AnswerFormat): number {
	return [...(format === 'rich' ? htmlText(body) : body)].length;
}

/** The format an answer to this kind of question is stored in. */
export function formatFor(question: FormQuestion): AnswerFormat {
	return question.type === 'long_text' ? 'rich' : 'plain';
}

const escapeHtml = (s: string) =>
	s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

/**
 * Plain text as HTML: blank-line-separated paragraphs, single newlines as <br>, and everything
 * escaped, so "a < b" stays text instead of being read as markup.
 */
export function plainToHtml(text: string): string {
	return text
		.split(/\n{2,}/)
		.map((p) => p.trim())
		.filter(Boolean)
		.map((p) => `<p>${escapeHtml(p).replace(/\n/g, '<br>')}</p>`)
		.join('');
}

/**
 * Why an answer can't be saved, in words for the artist, or null if it can. The server rejects a
 * whole autosave batch if any one answer is invalid, so the form uses this to keep a bad answer
 * local (and visibly flagged) instead of letting it block every other answer from saving.
 */
export function answerProblem(question: FormQuestion, body: string): string | null {
	const format = formatFor(question);
	if (isBlankAnswer(body, format)) return null; // blank clears the answer; required-ness is checked on submit
	if (answerLength(body, format) > question.max_length) {
		return `This is over the ${question.max_length.toLocaleString('en')}-character limit.`;
	}
	if (question.type === 'link' && !safeHttpUrl(body)) {
		return 'Enter a full link starting with http:// or https://.';
	}
	if (question.type === 'choice' && !question.options?.some((o) => o.value === body)) {
		return 'Choose one of the options.';
	}
	return null;
}

/** Keys of required questions whose current answer is blank, in question order. */
export function missingRequired(
	questions: FormQuestion[],
	current: Record<string, string | undefined>
): string[] {
	return [...questions]
		.sort((a, b) => a.position - b.position)
		.filter((q) => q.required && isBlankAnswer(current[q.key] ?? '', formatFor(q)))
		.map((q) => q.key);
}

/**
 * A short excerpt for cards and previews: whitespace collapsed, cut at a word boundary within `max`
 * characters, "…" added only when something was cut. Counts characters, not bytes, so Amharic is fine.
 */
export function truncateText(text: string, max = 140): string {
	const clean = text.replace(/\s+/g, ' ').trim();
	const chars = Array.from(clean);
	if (chars.length <= max) return clean;
	const cut = chars.slice(0, max).join('');
	const lastSpace = cut.lastIndexOf(' ');
	// Back up to the last whole word, unless that would throw away most of the excerpt.
	const body = lastSpace > max * 0.6 ? cut.slice(0, lastSpace) : cut;
	return `${body.replace(/[\s,.;:!?-]+$/, '')}…`;
}

