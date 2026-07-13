import { artPostService } from '$lib/application/artPosts';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const city = url.searchParams.get('city') ?? undefined;
	const medium = url.searchParams.get('medium') ?? undefined;
	const style = url.searchParams.get('style') ?? undefined;
	const yearRaw = url.searchParams.get('year');
	const year = yearRaw ? Number(yearRaw) : undefined;

	const posts = await artPostService
		.list({
			city,
			medium,
			style,
			year: Number.isFinite(year) ? year : undefined,
			limit: 48
		})
		.catch(() => []);

	const cities = [...new Set(posts.map((p) => p.city).filter(Boolean))].sort() as string[];
	const mediums = [...new Set(posts.map((p) => p.medium).filter(Boolean))].sort() as string[];
	const styles = [...new Set(posts.map((p) => p.style).filter(Boolean))].sort() as string[];
	const years = [...new Set(posts.map((p) => p.year).filter((y): y is number => y != null))].sort(
		(a, b) => b - a
	);

	return {
		source: 'api' as const,
		posts,
		cities,
		mediums,
		styles,
		years,
		residencies: [] as string[],
		exhibitions: [] as string[],
		filters: { city, medium, style, year }
	};
};
