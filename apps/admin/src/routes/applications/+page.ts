import { onboardingService } from '$lib/application/onboarding';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
  const applications = await onboardingService.listPending();
  return { applications };
};
