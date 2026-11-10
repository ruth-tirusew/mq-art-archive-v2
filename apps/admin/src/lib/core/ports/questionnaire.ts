import type {
  NudgePreview,
  QuestionnaireRevision,
  ReviewView,
  SaveDraftInput
} from '$lib/core/domain/questionnaire';

export interface QuestionnairePort {
  /** The working draft, or null when there is none. */
  getDraft(): Promise<QuestionnaireRevision | null>;
  /** Returns the existing draft or creates one cloned from the latest published revision. */
  startDraft(): Promise<QuestionnaireRevision>;
  saveDraft(input: SaveDraftInput): Promise<QuestionnaireRevision>;
  discardDraft(): Promise<void>;
  publish(): Promise<QuestionnaireRevision>;
  /** Published revisions, newest first, without their questions. */
  listRevisions(): Promise<QuestionnaireRevision[]>;
  getRevision(id: string): Promise<QuestionnaireRevision>;
  /** Null when nothing has been published yet. */
  previewNudge(): Promise<NudgePreview | null>;
  sendNudge(): Promise<number>;
  /** An application's answers, or null when no questionnaire answers exist for it. */
  getApplicationAnswers(applicationId: string): Promise<ReviewView | null>;
}
