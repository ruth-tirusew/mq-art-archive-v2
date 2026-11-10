import type {
	AnswerInput,
	PublicAnswer,
	QuestionnaireForm,
	QuestionnaireNudge
} from '$lib/core/domain/questionnaire';

export interface QuestionnairePort {
	listArtistAnswers(slug: string): Promise<PublicAnswer[]>;
	/** Null before any questionnaire is published. */
	getForm(): Promise<QuestionnaireForm | null>;
	saveAnswers(answers: AnswerInput[], options?: { keepalive?: boolean }): Promise<void>;
	getNudge(): Promise<QuestionnaireNudge | null>;
	dismissNudge(revisionId: string): Promise<void>;
}
