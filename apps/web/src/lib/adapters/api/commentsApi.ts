import { apiFetch } from '$lib/adapters/api/client';
import type { ActivityComment } from '$lib/core/domain/activity';
import type { Comment } from '$lib/core/domain/comment';

export interface CreateCommentInput {
	body: string;
	isGeneral: boolean;
	quotedText?: string;
	start?: number;
	end?: number;
}

export class CommentsApi {
	create(articleId: string, input: CreateCommentInput): Promise<Comment> {
		return apiFetch<Comment>(`/api/v1/wiki/articles/${articleId}/comments`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({
				body: input.body,
				is_general: input.isGeneral,
				quoted_text: input.quotedText,
				start: input.start,
				end: input.end
			})
		});
	}

	delete(commentId: string): Promise<void> {
		return apiFetch<void>(`/api/v1/wiki/comments/${commentId}`, { method: 'DELETE' });
	}

	list(articleId: string): Promise<Comment[]> {
		return apiFetch<Comment[]>(`/api/v1/wiki/articles/${articleId}/comments`);
	}

	listMyActivity(): Promise<ActivityComment[]> {
		return apiFetch<ActivityComment[]>('/api/v1/me/comments');
	}
}
