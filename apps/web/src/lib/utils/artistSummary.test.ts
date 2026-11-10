import { describe, expect, it } from 'vitest';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { PublicAnswer } from '$lib/core/domain/questionnaire';
import { summarizeArtist } from './artistSummary';

const artist = { id: '1', slug: 'a', display_name: 'A' } as ArtistProfile;
const answer = (key: string, body: string, format: 'plain' | 'rich' = 'plain'): PublicAnswer =>
	({ key, type: 'short_text', prompt_en: key, body, format, lang: 'en' }) as PublicAnswer;

describe('summarizeArtist', () => {
	it('shows nothing for an artist with no profile fields and no answers', () => {
		expect(summarizeArtist(artist, [])).toEqual({ tagline: '', bio: '', bioHtml: '', tags: [], facts: [] });
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

	it('reads a rich-text bio as plain words for previews, and keeps the HTML for the full bio', () => {
		const s = summarizeArtist({ ...artist, bio: '<h2>About</h2><p>I paint <strong>light</strong> &amp; shadow.</p>' }, []);
		expect(s.bio).toBe('About I paint light & shadow.');
		expect(s.bioHtml).toBe('<h2>About</h2><p>I paint <strong>light</strong> &amp; shadow.</p>');
		expect(s.tagline).toBe(''); // the bio is shown once, in About, so it never doubles as a tagline
	});

	it('shows the short-bio answer in About when the artist has not written a profile bio', () => {
		const rich = summarizeArtist(artist, [answer('short_bio', '<p>From <em>Harar</em>.</p>', 'rich')]);
		expect(rich.bioHtml).toBe('<p>From <em>Harar</em>.</p>');
		expect(rich.tagline).toBe(''); // and it is not repeated as a tagline either

		const plain = summarizeArtist(artist, [answer('short_bio', 'Born in <Harar>\nPaints.')]);
		expect(plain.bioHtml).toBe('<p>Born in &lt;Harar&gt;<br>Paints.</p>'); // escaped, never read as markup
	});

	it('prefers the profile bio over the short-bio answer', () => {
		const s = summarizeArtist({ ...artist, bio: '<p>Mine</p>' }, [answer('short_bio', 'Theirs')]);
		expect(s.bioHtml).toBe('<p>Mine</p>');
	});
});

