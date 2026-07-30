import { describe, expect, it } from 'vitest';
import { validateImageFile } from './media';

describe('validateImageFile', () => {
  it('rejects unsupported image types', () => {
    expect(validateImageFile({ type: 'image/gif', size: 100 })).toMatch(/JPEG, PNG, or WebP/);
  });

  it('rejects files larger than 10 MB', () => {
    expect(validateImageFile({ type: 'image/jpeg', size: 10 * 1024 * 1024 + 1 })).toMatch(
      /10 MB/
    );
  });

  it('accepts supported images up to 10 MB', () => {
    expect(validateImageFile({ type: 'image/webp', size: 10 * 1024 * 1024 })).toBeNull();
  });
});
