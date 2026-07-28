import { onboardingService } from '$lib/application/onboarding';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
  const applications = await requireAdmin(() => onboardingService.listPending());
  return { applications };
};
