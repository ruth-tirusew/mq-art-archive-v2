import { ApiError, apiFetch } from '$lib/adapters/api/client';
import type {
	AnswerInput,
	PublicAnswer,
	QuestionnaireForm,
	QuestionnaireNudge
} from '$lib/core/domain/questionnaire';

function json(method: 'PUT' | 'POST', body?: unknown, keepalive = false): RequestInit {
	return {
		method,
		keepalive,
		headers: { 'Content-Type': 'application/json' },
		body: body === undefined ? undefined : JSON.stringify(body)
	};
}

/** Resolves to null on 404 and rethrows anything else. */
async function orNull<T>(request: Promise<T>): Promise<T | null> {
	try {
		return await request;
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) return null;
		throw err;
	}
}

export class QuestionnaireApi {
	/** An artist's visible answers, in question order. 404 unless the profile is public. */
	listArtistAnswers(slug: string): Promise<PublicAnswer[]> {
		return apiFetch<PublicAnswer[]>(`/api/v1/artists/${encodeURIComponent(slug)}/answers`);
	}

	/** The signed-in user's form, or null before any questionnaire is published. */
	getForm(): Promise<QuestionnaireForm | null> {
		return orNull(apiFetch<QuestionnaireForm>('/api/v1/me/questionnaire'));
	}

	/**
	 * Saves the given answers together: all are stored or none are. `keepalive` lets the request
	 * finish after the page has been hidden or closed.
	 */
	saveAnswers(answers: AnswerInput[], options: { keepalive?: boolean } = {}): Promise<void> {
		return apiFetch<void>('/api/v1/me/questionnaire/answers', json('PUT', { answers }, options.keepalive));
	}

	getNudge(): Promise<QuestionnaireNudge | null> {
		return orNull(apiFetch<QuestionnaireNudge>('/api/v1/me/questionnaire/nudge'));
	}

	dismissNudge(revisionId: string): Promise<void> {
		return apiFetch<void>(
			`/api/v1/me/questionnaire/nudges/${encodeURIComponent(revisionId)}/dismiss`,
			json('POST')
		);
	}
}
