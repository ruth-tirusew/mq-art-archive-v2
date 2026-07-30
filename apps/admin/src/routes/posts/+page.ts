import { postsService } from '$lib/application/posts';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') as 'draft' | 'published' | 'archived' | null;
  const posts = await requireAdmin(() => postsService.list(status ?? undefined));
  return { posts, status: status ?? 'all' };
};
