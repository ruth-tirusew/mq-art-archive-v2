import { describe, expect, it } from 'vitest';
import { parseApplicationStatusFilter } from '$lib/core/domain/onboarding';

describe('parseApplicationStatusFilter', () => {
  it('accepts the four filters', () => {
    for (const status of ['pending', 'approved', 'rejected', 'all'] as const) {
      expect(parseApplicationStatusFilter(status)).toBe(status);
    }
  });
  it('falls back to the review queue for anything else, so a hand-edited URL still works', () => {
    expect(parseApplicationStatusFilter(null)).toBe('pending');
    expect(parseApplicationStatusFilter('')).toBe('pending');
    expect(parseApplicationStatusFilter('bogus')).toBe('pending');
  });
});
