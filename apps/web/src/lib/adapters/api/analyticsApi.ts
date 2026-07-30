import { apiFetch } from '$lib/adapters/api/client';
import type { AnalyticsView } from '$lib/core/domain/analytics';

export class AnalyticsApi {
	record(entityType: string, entityId: string): Promise<{ recorded: boolean }> {
		return apiFetch('/api/v1/analytics/view', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ entity_type: entityType, entity_id: entityId })
		});
	}

	query(entityType: string, entityId: string): Promise<AnalyticsView[]> {
		const query = new URLSearchParams({ entity_type: entityType, entity_id: entityId });
		return apiFetch<AnalyticsView[]>(`/api/v1/me/analytics?${query}`);
	}
}
