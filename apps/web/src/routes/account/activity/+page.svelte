<script lang="ts">
	import { onMount } from 'svelte';
	import { favoritesService } from '$lib/application/favorites';
	import { highlightsService } from '$lib/application/highlights';
	import { commentsService } from '$lib/application/comments';
	import type { Article } from '$lib/core/domain/content';
	import type { ActivityHighlight, ActivityComment } from '$lib/core/domain/activity';

	let loading = $state(true);
	let error = $state('');
	let favorites = $state<Article[]>([]);
	let highlights = $state<ActivityHighlight[]>([]);
	let comments = $state<ActivityComment[]>([]);

	onMount(async () => {
		try {
			[favorites, highlights, comments] = await Promise.all([
				favoritesService.listMine(),
				highlightsService.listMyActivity(),
				commentsService.listMyActivity()
			]);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load activity';
		} finally {
			loading = false;
		}
	});

	function formatTime(iso: string) {
		try {
			return new Date(iso).toLocaleDateString('en-GB', { day: '2-digit', month: 'short', year: 'numeric' });
		} catch {
			return iso;
		}
	}
</script>

<svelte:head><title>My activity — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl space-y-14 px-6 py-14 md:px-10 md:py-20">
	<div>
		<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Account</p>
		<h1 class="mt-4 font-display text-4xl text-foreground">My activity</h1>
	</div>

	{#if loading}
		<p class="text-muted-foreground">Loading…</p>
	{:else if error}
		<p class="text-sm text-destructive" role="alert">{error}</p>
	{:else}
		<div>
			<h2 class="font-display text-xl text-foreground">
				Saved articles {#if favorites.length > 0}({favorites.length}){/if}
			</h2>
			{#if favorites.length === 0}
				<p class="mt-3 text-sm text-muted-foreground">
					Nothing saved yet. <a href="/wiki" class="underline underline-offset-4">Browse the wiki</a>.
				</p>
			{:else}
				<ul class="mt-4 divide-y divide-border/60 border-y border-border/60">
					{#each favorites as article (article.id)}
						<li class="py-3">
							<a href="/wiki/{article.slug}" class="text-sm text-foreground underline-offset-4 hover:underline">
								{article.title}
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<div>
			<h2 class="font-display text-xl text-foreground">
				Highlights {#if highlights.length > 0}({highlights.length}){/if}
			</h2>
			{#if highlights.length === 0}
				<p class="mt-3 text-sm text-muted-foreground">No highlights yet.</p>
			{:else}
				<ul class="mt-4 space-y-4">
					{#each highlights as h (h.id)}
						<li class="border-l-2 border-accent/40 pl-4">
							<p class="text-sm italic text-foreground/90">"{h.quoted_text}"</p>
							<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
								<a href="/wiki/{h.article_slug}" class="text-accent hover:underline">{h.article_title}</a>
								· {formatTime(h.created_at)}
							</p>
						</li>
					{/each}
				</ul>
			{/if}
		</div>

		<div>
			<h2 class="font-display text-xl text-foreground">
				Comments {#if comments.length > 0}({comments.length}){/if}
			</h2>
			{#if comments.length === 0}
				<p class="mt-3 text-sm text-muted-foreground">No comments yet.</p>
			{:else}
				<ul class="mt-4 space-y-4">
					{#each comments as c (c.id)}
						<li class="border-l-2 border-border/60 pl-4">
							<p class="text-sm text-foreground/90">{c.body}</p>
							<p class="mt-1 font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
								<a href="/wiki/{c.article_slug}" class="text-accent hover:underline">{c.article_title}</a>
								· {c.is_general ? 'response' : 'comment'} · {formatTime(c.created_at)}
							</p>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</section>
