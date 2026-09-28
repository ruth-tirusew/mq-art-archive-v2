<script lang="ts">
	import type { Comment } from '$lib/core/domain/comment';

	let {
		comments,
		placeholder = 'Add to the discussion…',
		onSubmit,
		onDelete,
		currentUserId,
		signInHref = '/login'
	}: {
		comments: Comment[];
		placeholder?: string;
		onSubmit: (body: string) => Promise<void>;
		onDelete?: (commentId: string) => Promise<void>;
		currentUserId?: string;
		signInHref?: string;
	} = $props();

	let draft = $state('');
	let submitting = $state(false);
	let deletingId = $state('');
	let error = $state('');

	function formatTime(iso: string) {
		try {
			return new Date(iso).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
		} catch {
			return iso;
		}
	}

	async function submit() {
		const body = draft.trim();
		if (!body || submitting) return;
		submitting = true;
		error = '';
		try {
			await onSubmit(body);
			draft = '';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not post comment';
		} finally {
			submitting = false;
		}
	}

	async function remove(commentId: string) {
		if (!onDelete || deletingId) return;
		deletingId = commentId;
		try {
			await onDelete(commentId);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not delete comment';
		} finally {
			deletingId = '';
		}
	}
</script>

<div class="space-y-4" data-testid="web-wiki-comment-thread">
	{#if comments.length > 0}
		<ul class="space-y-4">
			{#each comments as comment (comment.id)}
				<li class="border-l-2 border-border/60 pl-4">
					<div class="flex items-center justify-between gap-2">
						<span class="font-mono text-[10px] uppercase tracking-[0.15em] text-foreground">
							{comment.author_name}
						</span>
						<div class="flex items-center gap-2">
							<span class="font-mono text-[9px] uppercase tracking-[0.15em] text-muted-foreground">
								{formatTime(comment.created_at)}
							</span>
							{#if onDelete && currentUserId && comment.user_id === currentUserId}
								<button
									type="button"
									disabled={deletingId === comment.id}
									onclick={() => remove(comment.id)}
									class="font-mono text-[9px] uppercase tracking-[0.15em] text-destructive underline underline-offset-2 disabled:opacity-50"
								>
									Delete
								</button>
							{/if}
						</div>
					</div>
					<p class="mt-1 text-sm leading-relaxed text-foreground/90">{comment.body}</p>
				</li>
			{/each}
		</ul>
	{/if}

	{#if currentUserId}
		<form
			class="flex items-start gap-2"
			onsubmit={(e) => {
				e.preventDefault();
				void submit();
			}}
		>
			<textarea
				bind:value={draft}
				{placeholder}
				rows={2}
				class="flex-1 rounded-sm border border-input bg-background px-3 py-2 text-sm outline-none focus-visible:border-ring"
			></textarea>
			<button
				type="submit"
				disabled={submitting || !draft.trim()}
				class="shrink-0 rounded-sm bg-foreground px-4 py-2 font-mono text-[10px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
			>
				{submitting ? '…' : 'Post'}
			</button>
		</form>
		{#if error}<p class="text-xs text-destructive" role="alert">{error}</p>{/if}
	{:else}
		<p class="text-xs text-muted-foreground">
			<a href={signInHref} class="text-foreground underline underline-offset-4">Sign in</a> to join the discussion.
		</p>
	{/if}
</div>
