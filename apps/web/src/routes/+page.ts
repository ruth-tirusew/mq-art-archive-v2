import type { EditorialSpread } from '$lib/components/home/editorialCompositions';
import type { PageLoad } from './$types';

function parseSpreads(payload: string | undefined): EditorialSpread[] {
	if (!payload) return [];
	try {
		const parsed: unknown = JSON.parse(payload);
		return Array.isArray(parsed) ? (parsed as EditorialSpread[]) : [];
	} catch {
		return [];
	}
}

/** Parse hero spreads on the load boundary so SSR/hydration see a plain array (Svelte 5 page props quirk). */
export const load: PageLoad = ({ data }) => {
	const fromJson = parseSpreads(data.editorialSpreadsPayload);
	const heroSpreads =
		fromJson.length > 0 ? fromJson : (data.editorialSpreads ?? data.editorialWalls ?? []);

	return { heroSpreads };
};
