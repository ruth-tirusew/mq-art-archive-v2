import { describe, expect, it } from 'vitest';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { PublicAnswer } from '$lib/core/domain/questionnaire';
import { summarizeArtist } from './artistSummary';

const artist = { id: '1', slug: 'a', display_name: 'A' } as ArtistProfile;
const answer = (key: string, body: string, format: 'plain' | 'rich' = 'plain'): PublicAnswer =>
	({ key, type: 'short_text', prompt_en: key, body, format, lang: 'en' }) as PublicAnswer;

describe('summarizeArtist', () => {
	it('shows nothing for an artist with no profile fields and no answers', () => {
		expect(summarizeArtist(artist, [])).toEqual({ tagline: '', bio: '', tags: [], facts: [] });
	});

	it('fills the tagline, bio, tags and languages from public answers', () => {
		const s = summarizeArtist(artist, [
			answer('tagline', 'Paint and memory'),
			answer('short_bio', '<p>Born in <b>Harar</b>.</p>', 'rich'),
			answer('tag_materials', 'oil, ink,  oil'),
			answer('languages', 'Amharic, English')
		]);
		expect(s.tagline).toBe('Paint and memory');
		expect(s.bio).toBe('Born in Harar.');
		expect(s.tags).toEqual(['oil', 'ink']);
		expect(s.facts).toEqual([{ label: 'Languages', value: 'Amharic, English' }]);
	});

	it('prefers what the artist typed on their profile', () => {
		const s = summarizeArtist({ ...artist, tagline: 'Typed' }, [answer('tagline', 'Answered')]);
		expect(s.tagline).toBe('Typed');
	});
});
