// @ts-nocheck
import { onboardingService } from '$lib/application/onboarding';
import type { PageLoad } from './$types';

export const load = async () => {
  const applications = await onboardingService.listPending();
  return { applications };
};
;null as any as PageLoad;