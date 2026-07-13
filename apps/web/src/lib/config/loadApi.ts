import { useApi } from '$lib/config/dataSource';

/** Run an API loader; on failure return the static fallback instead of throwing. */
export async function loadWithApiFallback<TApi, TStatic>(
	apiLoad: () => Promise<TApi>,
	staticLoad: () => TStatic
): Promise<TApi | TStatic> {
	if (!useApi) return staticLoad();
	try {
		return await apiLoad();
	} catch (err) {
		console.warn('[api] loader failed, using static fallback:', err);
		return staticLoad();
	}
}
