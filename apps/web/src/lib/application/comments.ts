import { CommentsApi, type CreateCommentInput } from '$lib/adapters/api/commentsApi';
import type { Comment } from '$lib/core/domain/comment';

const api = new CommentsApi();

export const commentsService = {
	create(articleId: string, input: CreateCommentInput): Promise<Comment> {
		return api.create(articleId, input);
	},
	delete(commentId: string): Promise<void> {
		return api.delete(commentId);
	},
	list(articleId: string): Promise<Comment[]> {
		return api.list(articleId);
	}
};
