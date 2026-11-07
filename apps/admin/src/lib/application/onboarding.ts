import {
  APPLICATIONS_PAGE_SIZE,
  type ApplicationPage,
  type ApplicationStatusFilter,
  type ApprovalStatus,
  type OnboardingApplication
} from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export function createOnboardingService(port: OnboardingPort) {
  return {
    /** One page of applications, starting after `offset` already-loaded rows. */
    list(status: ApplicationStatusFilter, offset = 0): Promise<ApplicationPage> {
      return port.list(status, { limit: APPLICATIONS_PAGE_SIZE, offset });
    },
    /** The review queue, for places that count or summarise it (the API caps one request at 200). */
    async listPending(): Promise<OnboardingApplication[]> {
      return (await port.list('pending', { limit: 200, offset: 0 })).items;
    },
    getById(id: string): Promise<OnboardingApplication> {
      return port.getById(id);
    },
    review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication> {
      return port.review(id, status, notes);
    }
  };
}
