import type {
  ApplicationPage,
  ApplicationStatusFilter,
  ApprovalStatus,
  OnboardingApplication
} from '$lib/core/domain/onboarding';

export interface OnboardingPort {
  list(status: ApplicationStatusFilter, options: { limit: number; offset: number }): Promise<ApplicationPage>;
  getById(id: string): Promise<OnboardingApplication>;
  review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication>;
}
