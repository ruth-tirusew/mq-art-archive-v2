import { articlesService } from '$lib/application/articles';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
  requireAdmin(async () => ({ article: await articlesService.getById(params.id) }));
