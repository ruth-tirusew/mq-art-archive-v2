import { articleService } from '$lib/application/articles';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		const article = await articleService.getBySlug(params.slug);
		return { error: false as const, article };
	} catch {
		return { error: true as const, article: null };
	}
};
