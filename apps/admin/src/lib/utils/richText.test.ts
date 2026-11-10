import { describe, expect, it } from 'vitest';
import { htmlText, plainToHtml, safeHttpUrl } from './richText';

describe('rich text helpers', () => {
  it('reads the visible words of editor HTML', () => {
    expect(htmlText('<h2>About</h2><p>I paint <strong>light</strong> &amp; shadow.</p><ul><li>a</li><li>b</li></ul>')).toBe('About I paint light & shadow. a b');
    expect(htmlText('<p></p>')).toBe('');
  });

  it('turns plain text into escaped paragraphs', () => {
    expect(plainToHtml('a < b & c\nnext\n\nsecond')).toBe('<p>a &lt; b &amp; c<br>next</p><p>second</p>');
  });

  it('only accepts plain http(s) links', () => {
    expect(safeHttpUrl('https://example.com/a')).toBe('https://example.com/a');
    for (const bad of ['javascript:alert(1)', 'data:text/html,x', '//evil.test', 'example.com', 'https://u:p@example.com']) {
      expect(safeHttpUrl(bad)).toBeNull();
    }
  });
});
