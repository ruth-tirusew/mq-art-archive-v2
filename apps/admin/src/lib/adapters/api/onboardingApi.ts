import { apiFetch, apiFetchPage } from '$lib/adapters/api/client';
import type {
  ApplicationPage,
  ApplicationStatusFilter,
  ApprovalStatus,
  OnboardingApplication
} from '$lib/core/domain/onboarding';
import type { OnboardingPort } from '$lib/core/ports/onboarding';

export class OnboardingApi implements OnboardingPort {
  list(status: ApplicationStatusFilter, { limit, offset }: { limit: number; offset: number }): Promise<ApplicationPage> {
    return apiFetchPage<OnboardingApplication>(
      `/admin/v1/applications?status=${encodeURIComponent(status)}&limit=${limit}&offset=${offset}`
    );
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
