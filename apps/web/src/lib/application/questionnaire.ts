import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import type { PublicAnswer } from '$lib/core/domain/questionnaire';

const api = new QuestionnaireApi();

export const questionnaireService = {
	/**
	 * Answers are extra content on a profile, so a failure (API hiccup, 404) must never break the
	 * page: it just shows no Q&A.
	 */
	async listArtistAnswers(slug: string): Promise<PublicAnswer[]> {
		try {
			return await api.listArtistAnswers(slug);
		} catch {
			return [];
		}
	}
};
