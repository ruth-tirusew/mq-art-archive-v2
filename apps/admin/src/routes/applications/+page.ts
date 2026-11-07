import { onboardingService } from '$lib/composition';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
  const applications = await requireAdmin(() => onboardingService.listPending());
  return { applications };
};
