import { onboardingService } from '$lib/application/onboarding';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
  requireAdmin(async () => ({ application: await onboardingService.getById(params.id) }));
