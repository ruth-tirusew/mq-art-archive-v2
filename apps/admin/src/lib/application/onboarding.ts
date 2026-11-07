import type { ApprovalStatus, OnboardingApplication } from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export function createOnboardingService(port: OnboardingPort) {
  return {
    listPending(): Promise<OnboardingApplication[]> {
      return port.listPending();
    },
    getById(id: string): Promise<OnboardingApplication> {
      return port.getById(id);
    },
    review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication> {
      return port.review(id, status, notes);
    }
  };
}
