import { redirect } from '@sveltejs/kit';
import { ApiError } from '$lib/adapters/api/client';
import { articlesService } from '$lib/application/articles';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const article = await articlesService.getById(params.id);
    return { article };
  } catch (err) {
    if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
      redirect(302, '/login');
    }
    throw err;
  }
};
