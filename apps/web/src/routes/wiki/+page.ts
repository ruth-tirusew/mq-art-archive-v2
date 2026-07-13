import { articleService } from '$lib/application/articles';
import { loadWithApiFallback } from '$lib/config/loadApi';
import { listWikiArticles, wikiCategories } from '$lib/data/wiki';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const category = url.searchParams.get('category') ?? undefined;

	return loadWithApiFallback(
		async () => {
			const articles = await articleService.listPublished({ category, limit: 100 });
			const categories = [...new Set(articles.map((a) => a.category).filter(Boolean))].sort() as string[];
			return { source: 'api' as const, articles, categories, filterCategory: category };
		},
		() => {
			const articles = listWikiArticles();
			const filtered =
				category && category !== 'All'
					? articles.filter((a) => a.category === category)
					: articles;
			return {
				source: 'static' as const,
				articles: filtered,
				categories: wikiCategories.filter((c) => c !== 'All'),
				filterCategory: category
			};
		}
	);
};
