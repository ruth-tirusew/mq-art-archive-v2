// @ts-nocheck
import { onboardingService } from '$lib/application/onboarding';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load = async () => {
  const applications = await requireAdmin(() => onboardingService.listPending());
  return { applications };
};
;null as any as PageLoad;