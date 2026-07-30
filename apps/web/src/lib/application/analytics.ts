import { AnalyticsApi } from '$lib/adapters/api/analyticsApi';
import type { AnalyticsView } from '$lib/core/domain/analytics';

const api = new AnalyticsApi();

export function recordPageView(entityType: string, entityId: string): void {
	if (!entityId) return;
	void api.record(entityType, entityId).catch(() => {
		// Analytics must never interrupt page rendering.
	});
}

export const analyticsService = {
	query(entityType: string, entityId: string): Promise<AnalyticsView[]> {
		return api.query(entityType, entityId);
	}
};
