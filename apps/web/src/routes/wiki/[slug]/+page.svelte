<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import type { PageData } from './$types';
	import { recordPageView } from '$lib/application/analytics';
	import FavoriteButton from '$lib/components/wiki/FavoriteButton.svelte';
	import SelectionToolbar from '$lib/components/wiki/SelectionToolbar.svelte';
	import CommentThread from '$lib/components/wiki/CommentThread.svelte';
	import { paragraphElements, paragraphStartOffsets, type TextAnchor } from '$lib/utils/textAnchor';
	import { segmentParagraph, type Span } from '$lib/utils/textSegments';
	import { authService, currentUser } from '$lib/application/auth';
	import { highlightsService } from '$lib/application/highlights';
	import { commentsService } from '$lib/application/comments';
	import { ApiError } from '$lib/adapters/api/client';
	import type { Highlight, PopularHighlight } from '$lib/core/domain/highlight';
	import type { Comment } from '$lib/core/domain/comment';

	let { data }: { data: PageData } = $props();

	const article = $derived(data.article);
	const title = $derived(article?.title ?? 'Wiki unavailable');
	// Foundation for comments/highlights (both anchor to a span of this body — see
	// $lib/utils/textAnchor): each rendered paragraph carries the index a selection
	// inside it needs to resolve back to a global offset in article.body.
	const paragraphs = $derived(article?.body ? paragraphElements(article.body) : []);
	const paragraphOffsets = $derived(article?.body ? paragraphStartOffsets(article.body) : []);
	const signInHref = $derived(`/login?return_to=${encodeURIComponent($page.url.pathname)}`);
	let bodyEl: HTMLDivElement | undefined = $state();

	let popularHighlight = $state<PopularHighlight | null>(null);
	let myHighlights = $state<Highlight[]>([]);
	let comments = $state<Comment[]>([]);
	let expandedParagraph = $state<number | null>(null);

	let composer = $state<{ anchor: TextAnchor; top: number; left: number } | null>(null);
	let composerDraft = $state('');
	let composerBusy = $state(false);
	let composerError = $state('');

	onMount(async () => {
		if (!article?.id) return;
		recordPageView('article', article.id);
		await authService.load();
		await refreshEngagement();
	});

	async function refreshEngagement() {
		if (!article?.id) return;
		const [popular, list] = await Promise.all([
			highlightsService.getPopular(article.id).catch(() => null),
			commentsService.list(article.id).catch(() => [])
		]);
		popularHighlight = popular;
		comments = list;
		if ($currentUser) {
			myHighlights = await highlightsService.listMine(article.id).catch(() => []);
		}
	}

	function requireLogin() {
		void goto(`/login?return_to=${encodeURIComponent($page.url.pathname)}`);
	}

	async function handleHighlight(anchor: TextAnchor) {
		if (!article?.id) return;
		if (!$currentUser) return requireLogin();
		try {
			await highlightsService.create(article.id, anchor.quotedText, anchor.textOffsetStart, anchor.textOffsetEnd);
			myHighlights = await highlightsService.listMine(article.id);
		} catch (e) {
			if (e instanceof ApiError && e.status === 401) requireLogin();
		}
	}

	// Estimated composer height for the below/above-selection flip below — the exact
	// rendered height isn't known until after layout, but this is close enough (measured
	// ~160px in practice) to decide which side has room.
	const COMPOSER_HEIGHT_ESTIMATE = 180;
	const COMPOSER_WIDTH = 288; // matches the w-72 class below

	function handleCommentSelect(anchor: TextAnchor) {
		if (!$currentUser) return requireLogin();
		const selection = window.getSelection();
		const range = selection?.rangeCount ? selection.getRangeAt(0) : null;
		const rect = range?.getBoundingClientRect();
		// position: fixed is already viewport-relative — no window.scrollX/scrollY here
		// (see the same fix in SelectionToolbar.svelte for why).
		const below = (rect?.bottom ?? 0) + 8;
		const fitsBelow = below + COMPOSER_HEIGHT_ESTIMATE <= window.innerHeight;
		const top = fitsBelow ? below : Math.max(8, (rect?.top ?? 0) - COMPOSER_HEIGHT_ESTIMATE - 8);
		const left = Math.min(Math.max(8, rect?.left ?? 0), window.innerWidth - COMPOSER_WIDTH - 8);

		composer = { anchor, top, left };
		composerDraft = '';
		composerError = '';
	}

	async function submitAnchoredComment() {
		if (!article?.id || !composer || !composerDraft.trim()) return;
		composerBusy = true;
		composerError = '';
		try {
			await commentsService.create(article.id, {
				body: composerDraft.trim(),
				isGeneral: false,
				quotedText: composer.anchor.quotedText,
				start: composer.anchor.textOffsetStart,
				end: composer.anchor.textOffsetEnd
			});
			comments = await commentsService.list(article.id);
			composer = null;
		} catch (e) {
			composerError = e instanceof Error ? e.message : 'Could not post comment';
		} finally {
			composerBusy = false;
		}
	}

	async function submitGeneralComment(body: string) {
		if (!article?.id) return;
		if (!$currentUser) return requireLogin();
		await commentsService.create(article.id, { body, isGeneral: true });
		comments = await commentsService.list(article.id);
	}

	async function submitReply(anchoredTo: Comment, body: string) {
		if (!article?.id) return;
		await commentsService.create(article.id, {
			body,
			isGeneral: false,
			quotedText: anchoredTo.quoted_text,
			start: anchoredTo.start,
			end: anchoredTo.end
		});
		comments = await commentsService.list(article.id);
	}

	async function deleteComment(commentId: string) {
		await commentsService.delete(commentId);
		comments = comments.filter((c) => c.id !== commentId);
	}

	function paragraphSpans(globalStart: number, length: number) {
		const end = globalStart + length;
		const mine: Span[] = myHighlights
			.filter((h) => h.matched && h.start < end && h.end > globalStart)
			.map((h) => ({ start: h.start, end: h.end }));
		const popular: Span[] =
			popularHighlight && popularHighlight.start < end && popularHighlight.end > globalStart
				? [{ start: popularHighlight.start, end: popularHighlight.end }]
				: [];
		const anchoredComments = comments.filter(
			(c) => !c.is_general && c.matched && c.start < end && c.end > globalStart
		);
		const comment: Span[] = anchoredComments.map((c) => ({ start: c.start, end: c.end }));
		return { spansByTag: { mine, popular, comment }, anchoredComments };
	}

	function segmentClass(tags: string[]): string {
		const classes: string[] = [];
		if (tags.includes('popular')) classes.push('bg-amber-400/30');
		if (tags.includes('mine')) classes.push('bg-accent/25');
		if (tags.includes('comment')) classes.push('underline decoration-accent decoration-2 underline-offset-4');
		return classes.join(' ');
	}
</script>

<svelte:head>
	<title>{title} — Artiv Wiki</title>
	<meta name="description" content={article?.excerpt ?? 'A community-written guide for Ethiopian creatives.'} />
</svelte:head>

{#if article}
<article class="mx-auto max-w-[900px] px-6 py-16 md:px-10 md:py-24">
	<a
		href="/wiki"
		class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground hover:text-foreground"
	>
		← Back to wiki
	</a>

	<div class="mt-8 flex flex-wrap items-center justify-between gap-3">
		<div class="flex flex-wrap items-center gap-3">
			<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
				{article.category ?? 'General'}
			</span>
			{#if article.updated_at}
				<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
					· last edited {new Date(article.updated_at).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' })}
				</span>
			{/if}
			{#if popularHighlight}
				<span
					class="font-mono text-[9px] uppercase tracking-[0.15em] text-amber-600 dark:text-amber-400"
					title="The passage {popularHighlight.count} readers highlighted most"
				>
					★ {popularHighlight.count} readers highlighted a passage
				</span>
			{/if}
		</div>
		<FavoriteButton articleId={article.id} />
	</div>

	<h1 class="mt-5 font-display text-4xl leading-tight text-foreground md:text-5xl">{title}</h1>

	{#if article.excerpt}
		<p class="mt-6 text-lg leading-relaxed text-muted-foreground">{article.excerpt}</p>
	{/if}

	<div
		bind:this={bodyEl}
		class="relative mt-10 space-y-6 text-base leading-relaxed text-foreground/90 md:text-lg wiki-body"
	>
		{#each paragraphs as p, i (p.index)}
			{@const globalStart = paragraphOffsets[p.index] ?? 0}
			{@const { spansByTag, anchoredComments } = paragraphSpans(globalStart, p.text.length)}
			{@const segments = segmentParagraph(p.text, globalStart, spansByTag)}
			<div>
				<p data-paragraph-index={p.index}>
					{#each segments as seg, si (si)}
						<span class={segmentClass(seg.tags)}>{seg.text}</span>
					{/each}
				</p>
				{#if anchoredComments.length > 0}
					<button
						type="button"
						data-testid="web-wiki-comment-marker"
						onclick={() => (expandedParagraph = expandedParagraph === i ? null : i)}
						class="mt-2 inline-flex items-center gap-1 font-mono text-[10px] uppercase tracking-[0.15em] text-accent hover:underline"
					>
						💬 {anchoredComments.length}
						{anchoredComments.length === 1 ? 'comment' : 'comments'}
					</button>
					{#if expandedParagraph === i}
						<div class="mt-3 rounded-sm border border-border/60 bg-card/30 p-4">
							<CommentThread
								comments={anchoredComments}
								placeholder="Reply to this thread…"
								currentUserId={$currentUser?.id}
								{signInHref}
								onSubmit={(body) => submitReply(anchoredComments[0], body)}
								onDelete={deleteComment}
							/>
						</div>
					{/if}
				{/if}
			</div>
		{/each}
	</div>

	{#if bodyEl && article.body}
		<SelectionToolbar container={bodyEl} body={article.body} onHighlight={handleHighlight} onComment={handleCommentSelect} />
	{/if}

	{#if composer}
		<div
			class="fixed z-50 w-72 rounded-sm border border-border bg-card p-3 shadow-lg"
			style="top: {composer.top}px; left: {composer.left}px;"
		>
			<p class="mb-2 line-clamp-2 border-l-2 border-accent pl-2 text-xs italic text-muted-foreground">
				"{composer.anchor.quotedText}"
			</p>
			<textarea
				bind:value={composerDraft}
				rows={3}
				placeholder="Add a comment…"
				class="w-full rounded-sm border border-input bg-background px-2 py-1.5 text-sm outline-none"
			></textarea>
			{#if composerError}<p class="mt-1 text-xs text-destructive" role="alert">{composerError}</p>{/if}
			<div class="mt-2 flex justify-end gap-2">
				<button
					type="button"
					onclick={() => (composer = null)}
					class="font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground"
				>
					Cancel
				</button>
				<button
					type="button"
					disabled={composerBusy || !composerDraft.trim()}
					onclick={submitAnchoredComment}
					class="rounded-sm bg-foreground px-3 py-1 font-mono text-[10px] uppercase tracking-[0.15em] text-background disabled:opacity-50"
				>
					{composerBusy ? '…' : 'Comment'}
				</button>
			</div>
		</div>
	{/if}

	<section class="mt-16 border-t border-border/60 pt-10">
		<h2 class="font-display text-2xl text-foreground">
			Responses {#if comments.filter((c) => c.is_general).length > 0}({comments.filter((c) => c.is_general).length}){/if}
		</h2>
		<div class="mt-6">
			<CommentThread
				comments={comments.filter((c) => c.is_general)}
				currentUserId={$currentUser?.id}
				{signInHref}
				onSubmit={submitGeneralComment}
				onDelete={deleteComment}
			/>
		</div>
	</section>
</article>
{:else}
	<section class="mx-auto max-w-[900px] px-6 py-24 text-center md:px-10">
		<h1 class="font-display text-4xl text-foreground">Wiki entry unavailable</h1>
		<p class="mt-4 text-muted-foreground">We couldn't load this entry from the API.</p>
		<button
			type="button"
			onclick={() => location.reload()}
			class="mt-6 rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background"
		>
			Retry
		</button>
	</section>
{/if}

<style>
	.wiki-body :global(p) {
		margin: 0;
	}
</style>
