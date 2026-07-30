import { redirect } from '@sveltejs/kit';
import { ApiError } from '$lib/adapters/api/client';
import { articlesService } from '$lib/application/articles';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
  const status = url.searchParams.get('status') as 'draft' | 'published' | 'archived' | null;
  try {
    const articles = await articlesService.list(status ?? undefined);
    return { articles, status: status ?? 'all' };
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect(302, '/login');
    }
    throw err;
  }
};
