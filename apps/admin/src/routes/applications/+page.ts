import { onboardingService } from '$lib/composition';
import { parseApplicationStatusFilter } from '$lib/core/domain/onboarding';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = parseApplicationStatusFilter(url.searchParams.get('status'));
  const page = await requireAdmin(() => onboardingService.list(status));
  return { status, applications: page.items, total: page.total };
};
