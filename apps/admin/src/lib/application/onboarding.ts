import {
  APPLICATIONS_PAGE_SIZE,
  type ApplicationPage,
  type ApplicationStatusFilter,
  type ApprovalStatus,
  type OnboardingApplication
} from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

/** The API's per-request maximum, and a bound so a misbehaving total can never loop forever. */
const QUEUE_PAGE_SIZE = 200;
const MAX_QUEUE_PAGES = 50;

export function createOnboardingService(port: OnboardingPort) {
  return {
    /** One page of applications, starting after `offset` already-loaded rows. */
    list(status: ApplicationStatusFilter, offset = 0): Promise<ApplicationPage> {
      return port.list(status, { limit: APPLICATIONS_PAGE_SIZE, offset });
    },
    /**
     * The whole review queue, for places that count or chart it. The API caps one request at 200,
     * so this pages until it has everything the API says exists (X-Total-Count), instead of
     * silently stopping at the first 200 and undercounting.
     */
    async listPending(): Promise<OnboardingApplication[]> {
      const items: OnboardingApplication[] = [];
      for (let guard = 0; guard < MAX_QUEUE_PAGES; guard++) {
        const page = await port.list('pending', { limit: QUEUE_PAGE_SIZE, offset: items.length });
        items.push(...page.items);
        if (page.items.length === 0 || items.length >= page.total) break;
      }
      return items;
    },
    getById(id: string): Promise<OnboardingApplication> {
      return port.getById(id);
    },
    review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication> {
      return port.review(id, status, notes);
    }
  };
}
