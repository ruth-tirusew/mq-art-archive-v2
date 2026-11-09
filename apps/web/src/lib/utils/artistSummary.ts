import type { PublicAnswer } from '$lib/core/domain/questionnaire';
import type { ArtistProfile } from '$lib/core/domain/profile';
import { answerText, htmlText } from '$lib/utils/answers';
import { artistBorn, artistDiscipline, artistInfluences, artistLocation, artistYearsActive } from '$lib/utils/fields';

/** Answers shown in the page header, or consents and credit settings that are not content. */
export const HEADER_ANSWER_KEYS = /^(tagline|short_bio|languages|credit_|consent_|tag_)/;

/** Plain text of one public answer, or '' if the artist did not give it. */
function plainAnswer(answers: PublicAnswer[], key: string): string {
	const a = answers.find((x) => x.key === key);
	if (!a) return '';
	return (a.format === 'rich' ? htmlText(a.body) : answerText(a)).trim();
}

function splitTags(text: string): string[] {
	return text
		.split(/[,\n]/)
		.map((t) => t.trim())
		.filter(Boolean);
}

export type ArtistSummary = {
	tagline: string;
	/** A longer blurb: the profile bio, else the short bio answer. */
	bio: string;
	tags: string[];
	facts: { label: string; value: string }[];
};

/**
 * What a profile shows about an artist. Onboarding collects a tagline, short bio, tags and
 * languages as questionnaire answers rather than profile fields, so this reads the public answers
 * and falls back to the profile fields an artist may have filled in later.
 */
export function summarizeArtist(artist: ArtistProfile, answers: PublicAnswer[]): ArtistSummary {
	const bio = artist.bio || plainAnswer(answers, 'short_bio');
	const tagline = artist.tagline || plainAnswer(answers, 'tagline') || bio;
	const tags = [
		...artistInfluences(artist),
		...splitTags(plainAnswer(answers, 'tag_influences')),
		...splitTags(plainAnswer(answers, 'tag_materials')),
		...splitTags(plainAnswer(answers, 'tag_techniques')),
		...splitTags(plainAnswer(answers, 'tag_collectives'))
	].filter((t, i, all) => all.indexOf(t) === i);

	const facts = [
		{ label: 'Born', value: artistBorn(artist) ?? '' },
		{ label: 'Based', value: artistLocation(artist) },
		{ label: 'Discipline', value: artistDiscipline(artist) ?? '' },
		{ label: 'Active', value: artistYearsActive(artist) ?? '' },
		{ label: 'Languages', value: plainAnswer(answers, 'languages') }
	].filter((f) => f.value);

	return { tagline, bio, tags, facts };
}
