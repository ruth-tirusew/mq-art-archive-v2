import { articleService } from '$lib/application/articles';
import { getWikiArticle } from '$lib/data/wiki';
import { useApi } from '$lib/config/dataSource';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	if (!useApi) {
		const article = getWikiArticle(params.slug);
		if (!article) error(404, 'Article not found');
		return { source: 'static' as const, article };
	}

	try {
		const article = await articleService.getBySlug(params.slug);
		return { source: 'api' as const, article };
	} catch {
		const article = getWikiArticle(params.slug);
		if (article) return { source: 'static' as const, article };
		error(404, 'Article not found');
	}
};
