<script lang="ts">
	import { onMount } from 'svelte';
	import { artPostService } from '$lib/application/artPosts';
	import type { ArtPost } from '$lib/core/domain/art';

	type StatusFilter = 'all' | 'draft' | 'published' | 'archived';

	let posts = $state<ArtPost[]>([]);
	let loading = $state(true);
	let error = $state('');
	let filter = $state<StatusFilter>('all');
	let acting = $state<string | null>(null);

	const filtered = $derived(
		filter === 'all' ? posts : posts.filter((p) => (p.status ?? 'draft') === filter)
	);

	onMount(async () => {
		await loadPosts();
	});

	async function loadPosts() {
		loading = true;
		error = '';
		try {
			posts = await artPostService.listMyPosts();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load posts';
		} finally {
			loading = false;
		}
	}

	async function publish(id: string) {
		acting = id;
		try {
			const updated = await artPostService.publishMyPost(id);
			posts = posts.map((p) => (p.id === id ? updated : p));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Publish failed';
		} finally {
			acting = null;
		}
	}

	async function archive(id: string) {
		acting = id;
		try {
			const updated = await artPostService.archiveMyPost(id);
			posts = posts.map((p) => (p.id === id ? updated : p));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Archive failed';
		} finally {
			acting = null;
		}
	}
</script>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<div class="flex flex-wrap items-end justify-between gap-4">
		<div>
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Studio · Posts</p>
			<h1 class="mt-4 font-display text-4xl text-foreground">Your work</h1>
			<p class="mt-3 text-sm text-muted-foreground">Draft, publish, and archive pieces in your file.</p>
		</div>
		<a
			href="/studio/posts/new"
			class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90"
		>
			New draft
		</a>
	</div>

	<div class="mt-8 flex flex-wrap gap-2 font-mono text-[10px] uppercase tracking-[0.2em]">
		{#each ['all', 'draft', 'published', 'archived'] as f}
			<button
				type="button"
				onclick={() => (filter = f as StatusFilter)}
				class="rounded-full border px-3 py-1 transition {filter === f
					? 'border-foreground bg-foreground text-background'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				{f}
			</button>
		{/each}
	</div>

	{#if error}
		<p class="mt-6 text-sm text-destructive" role="alert">{error}</p>
	{/if}

	<div class="mt-8">
		{#if loading}
			<p class="text-muted-foreground">Loading…</p>
		{:else if filtered.length === 0}
			<p class="text-sm text-muted-foreground">No posts in this view.</p>
		{:else}
			<ul class="divide-y divide-border/60 border-y border-border/60">
				{#each filtered as post}
					<li class="py-4">
						<div class="flex flex-wrap items-start justify-between gap-4">
							<div>
								<a href="/studio/posts/{post.id}" class="font-display text-lg text-foreground hover:text-accent">
									{post.title}
								</a>
								{#if post.medium}
									<p class="mt-1 text-sm text-muted-foreground">{post.medium}</p>
								{/if}
							</div>
							<div class="flex flex-wrap items-center gap-3">
								<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
									{post.status ?? 'draft'}
								</span>
								<a
									href="/studio/posts/{post.id}"
									class="font-mono text-[10px] uppercase tracking-[0.2em] text-foreground underline decoration-accent underline-offset-4"
								>
									Edit
								</a>
								{#if (post.status ?? 'draft') === 'draft' || post.status === 'archived'}
									<button
										type="button"
										class="font-mono text-[10px] uppercase tracking-[0.2em] text-accent disabled:opacity-50"
										disabled={acting === post.id}
										onclick={() => void publish(post.id)}
									>
										Publish
									</button>
								{/if}
								{#if post.status === 'published'}
									<button
										type="button"
										class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground disabled:opacity-50"
										disabled={acting === post.id}
										onclick={() => void archive(post.id)}
									>
										Archive
									</button>
								{/if}
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</section>
