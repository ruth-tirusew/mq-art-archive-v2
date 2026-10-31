import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import type {
	AnswerInput,
	PublicAnswer,
	QuestionnaireForm,
	QuestionnaireNudge
} from '$lib/core/domain/questionnaire';

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
	},

	/** Null means no questionnaire has been published yet, so the plain application applies. */
	getForm(): Promise<QuestionnaireForm | null> {
		return api.getForm();
	},
	saveAnswers(answers: AnswerInput[], options?: { keepalive?: boolean }): Promise<void> {
		return api.saveAnswers(answers, options);
	},
	/** A nudge is a convenience; failing to load one must never get in the artist's way. */
	async getNudge(): Promise<QuestionnaireNudge | null> {
		try {
			return await api.getNudge();
		} catch {
			return null;
		}
	},
	dismissNudge(revisionId: string): Promise<void> {
		return api.dismissNudge(revisionId);
	}
};
