import type { ApprovalStatus, OnboardingApplication } from '$lib/core/domain/onboarding';

export interface OnboardingPort {
  listPending(): Promise<OnboardingApplication[]>;
  getById(id: string): Promise<OnboardingApplication>;
  review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication>;
}
