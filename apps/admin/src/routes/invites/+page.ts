import { invitesService } from '$lib/composition';
import { parseInviteStatusFilter } from '$lib/core/domain/invite';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = parseInviteStatusFilter(url.searchParams.get('status'));
  const page = await requireAdmin(() => invitesService.list(status));
  return { status, invites: page.items, total: page.total };
};
