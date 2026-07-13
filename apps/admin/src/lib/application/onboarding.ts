import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import type { ApprovalStatus, OnboardingApplication } from '$lib/core/domain/onboarding';

const api = new OnboardingApi();

export const onboardingService = {
  listPending(): Promise<OnboardingApplication[]> {
    return api.listPending();
  },
  getById(id: string): Promise<OnboardingApplication> {
    return api.getById(id);
  },
  review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication> {
    return api.review(id, status, notes);
  }
};
