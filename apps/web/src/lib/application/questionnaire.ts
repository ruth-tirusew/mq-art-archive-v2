import type {
	AnswerInput,
	PublicAnswer,
	QuestionnaireForm,
	QuestionnaireNudge
} from '$lib/core/domain/questionnaire';
import type { QuestionnairePort } from '$lib/core/ports/questionnaire';

export function createQuestionnaireService(port: QuestionnairePort) {
	return {
		/**
		 * Answers are extra content on a profile, so a failure (API hiccup, 404) must never break the
		 * page: it just shows no Q&A.
		 */
		async listArtistAnswers(slug: string): Promise<PublicAnswer[]> {
			try {
				return await port.listArtistAnswers(slug);
			} catch {
				return [];
			}
		},

		/** Null means no questionnaire has been published yet, so the plain application applies. */
		getForm(): Promise<QuestionnaireForm | null> {
			return port.getForm();
		},
		saveAnswers(answers: AnswerInput[], options?: { keepalive?: boolean }): Promise<void> {
			return port.saveAnswers(answers, options);
		},
		/** A nudge is a convenience; failing to load one must never get in the artist's way. */
		async getNudge(): Promise<QuestionnaireNudge | null> {
			try {
				return await port.getNudge();
			} catch {
				return null;
			}
		},
		dismissNudge(revisionId: string): Promise<void> {
			return port.dismissNudge(revisionId);
		}
	};
}

export type QuestionnaireService = ReturnType<typeof createQuestionnaireService>;
