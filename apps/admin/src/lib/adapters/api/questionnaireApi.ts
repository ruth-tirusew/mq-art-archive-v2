import { ApiError, apiFetch, jsonRequest } from '$lib/adapters/api/client';
import type {
  NudgePreview,
  QuestionnaireRevision,
  ReviewView,
  SaveDraftInput
} from '$lib/core/domain/questionnaire';
import type { QuestionnairePort } from '$lib/core/ports/questionnaire';

/** Resolves to null when the API answers 404, rethrows everything else. */
async function orNull<T>(request: Promise<T>): Promise<T | null> {
  try {
    return await request;
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

export class QuestionnaireApi implements QuestionnairePort {
  getDraft(): Promise<QuestionnaireRevision | null> {
    return orNull(apiFetch<QuestionnaireRevision>('/admin/v1/questionnaire/draft'));
  }

  startDraft(): Promise<QuestionnaireRevision> {
    return apiFetch<QuestionnaireRevision>('/admin/v1/questionnaire/draft', jsonRequest('POST'));
  }

  saveDraft(input: SaveDraftInput): Promise<QuestionnaireRevision> {
    return apiFetch<QuestionnaireRevision>('/admin/v1/questionnaire/draft', jsonRequest('PUT', input));
  }

  discardDraft(): Promise<void> {
    return apiFetch<void>('/admin/v1/questionnaire/draft', { method: 'DELETE' });
  }

  publish(): Promise<QuestionnaireRevision> {
    return apiFetch<QuestionnaireRevision>('/admin/v1/questionnaire/draft/publish', jsonRequest('POST'));
  }

  listRevisions(): Promise<QuestionnaireRevision[]> {
    return apiFetch<QuestionnaireRevision[]>('/admin/v1/questionnaire/revisions');
  }

  getRevision(id: string): Promise<QuestionnaireRevision> {
    return apiFetch<QuestionnaireRevision>(`/admin/v1/questionnaire/revisions/${encodeURIComponent(id)}`);
  }

  previewNudge(): Promise<NudgePreview | null> {
    return orNull(apiFetch<NudgePreview>('/admin/v1/questionnaire/nudge'));
  }

  async sendNudge(): Promise<number> {
    const result = await apiFetch<{ sent: number }>('/admin/v1/questionnaire/nudge', jsonRequest('POST'));
    return result.sent;
  }

  getApplicationAnswers(applicationId: string): Promise<ReviewView | null> {
    return orNull(apiFetch<ReviewView>(`/admin/v1/applications/${encodeURIComponent(applicationId)}/answers`));
  }
}
