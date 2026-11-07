import type {
  NudgePreview,
  QuestionnaireRevision,
  ReviewView,
  SaveDraftInput
} from '$lib/core/domain/questionnaire';
import type { QuestionnairePort } from '$lib/core/ports/questionnaire';

export function createQuestionnaireService(port: QuestionnairePort) {
  return {
    getDraft(): Promise<QuestionnaireRevision | null> {
      return port.getDraft();
    },
    startDraft(): Promise<QuestionnaireRevision> {
      return port.startDraft();
    },
    saveDraft(input: SaveDraftInput): Promise<QuestionnaireRevision> {
      return port.saveDraft(input);
    },
    discardDraft(): Promise<void> {
      return port.discardDraft();
    },
    publish(): Promise<QuestionnaireRevision> {
      return port.publish();
    },
    listRevisions(): Promise<QuestionnaireRevision[]> {
      return port.listRevisions();
    },
    getRevision(id: string): Promise<QuestionnaireRevision> {
      return port.getRevision(id);
    },
    previewNudge(): Promise<NudgePreview | null> {
      return port.previewNudge();
    },
    sendNudge(): Promise<number> {
      return port.sendNudge();
    },
    getApplicationAnswers(applicationId: string): Promise<ReviewView | null> {
      return port.getApplicationAnswers(applicationId);
    }
  };
}
