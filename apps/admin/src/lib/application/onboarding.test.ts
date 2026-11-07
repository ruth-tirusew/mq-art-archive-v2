import { describe, expect, it } from 'vitest';
import { createOnboardingService } from '$lib/application/onboarding';
import type { OnboardingApplication } from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

const app = (n: number) => ({ id: `app-${n}`, status: 'pending' }) as OnboardingApplication;

function portWith(total: number, calls: { limit: number; offset: number }[]): OnboardingPort {
  return {
    async list(_status, { limit, offset }) {
      calls.push({ limit, offset });
      const items = Array.from({ length: Math.max(0, Math.min(limit, total - offset)) }, (_, i) => app(offset + i));
      return { items, total };
    },
    getById: async () => app(0),
    review: async () => app(0)
  };
}

describe('listPending', () => {
  it('pages through a queue longer than one request so the dashboard count is not capped', async () => {
    const calls: { limit: number; offset: number }[] = [];
    const all = await createOnboardingService(portWith(450, calls)).listPending();
    expect(all).toHaveLength(450);
    expect(calls.map((c) => c.offset)).toEqual([0, 200, 400]);
    expect(new Set(all.map((a) => a.id)).size).toBe(450);
  });

  it('makes a single request for a short queue, and none beyond an empty one', async () => {
    const calls: { limit: number; offset: number }[] = [];
    expect(await createOnboardingService(portWith(3, calls)).listPending()).toHaveLength(3);
    expect(calls).toHaveLength(1);
    expect(await createOnboardingService(portWith(0, [])).listPending()).toEqual([]);
  });

  it('stops if the server keeps reporting a bigger total than it will return', async () => {
    const calls: unknown[] = [];
    const port: OnboardingPort = {
      ...portWith(0, []),
      async list() {
        calls.push(1);
        return { items: [], total: 999 };
      }
    };
    expect(await createOnboardingService(port).listPending()).toEqual([]);
    expect(calls).toHaveLength(1);
  });
});
