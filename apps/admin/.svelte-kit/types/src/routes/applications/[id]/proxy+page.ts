// @ts-nocheck
import { onboardingService } from '$lib/application/onboarding';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load = ({ params }: Parameters<PageLoad>[0]) =>
  requireAdmin(async () => ({ application: await onboardingService.getById(params.id) }));
