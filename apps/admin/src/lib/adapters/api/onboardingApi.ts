import { apiFetch } from '$lib/adapters/api/client';
import type { ApprovalStatus, OnboardingApplication } from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export class OnboardingApi implements OnboardingPort {
  listPending(): Promise<OnboardingApplication[]> {
    return apiFetch<OnboardingApplication[]>('/admin/v1/applications');
  }

  getById(id: string): Promise<OnboardingApplication> {
    return apiFetch<OnboardingApplication>(`/admin/v1/applications/${id}`);
  }

  review(id: string, status: ApprovalStatus, notes: string): Promise<OnboardingApplication> {
    return apiFetch<OnboardingApplication>(`/admin/v1/applications/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ status, notes })
    });
  }
}
