import { articleService } from '$lib/application/articles';
import { guardLoad } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
	guardLoad(async () => ({ article: await articleService.getBySlug(params.slug) }), 'We could not find that entry.');
