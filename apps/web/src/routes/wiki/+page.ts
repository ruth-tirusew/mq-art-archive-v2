import { articleService } from '$lib/application/articles';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const category = url.searchParams.get('category') ?? undefined;

	try {
		const articles = await articleService.listPublished({ category, limit: 100 });
		const categories = [...new Set(articles.map((a) => a.category).filter(Boolean))].sort() as string[];
		return { error: false as const, articles, categories, filterCategory: category };
	} catch {
		return {
			error: true as const,
			articles: [],
			categories: [],
			filterCategory: category
		};
	}
};
