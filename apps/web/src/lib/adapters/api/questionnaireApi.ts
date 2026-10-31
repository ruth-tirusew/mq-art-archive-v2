import { apiFetch } from '$lib/adapters/api/client';
import type { PublicAnswer } from '$lib/core/domain/questionnaire';

export class QuestionnaireApi {
	/** An artist's visible answers, in question order. 404 unless the profile is public. */
	listArtistAnswers(slug: string): Promise<PublicAnswer[]> {
		return apiFetch<PublicAnswer[]>(`/api/v1/artists/${encodeURIComponent(slug)}/answers`);
	}
}
