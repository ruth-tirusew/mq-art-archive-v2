import { articlesService } from '$lib/application/articles';
import { requireAdmin } from '$lib/utils/loadGuard';
import { PAGE_SIZE } from '$lib/utils/paging';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') as 'draft' | 'published' | 'archived' | null;
  const articles = await requireAdmin(() => articlesService.list(status ?? undefined, { limit: PAGE_SIZE }));
  return { articles, status: status ?? 'all' };
};
