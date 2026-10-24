<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { onMount } from 'svelte';
	import { artPostService } from '$lib/application/artPosts';
	import type { ArtPost } from '$lib/core/domain/art';
	import { Alert, Badge, Button, ConfirmDialog, createConfirm, EmptyState, ErrorState, PageHeader, toast } from '$ui';

	type StatusFilter = 'all' | 'draft' | 'published' | 'archived';

	let posts = $state<ArtPost[]>([]);
	let loading = $state(true);
	let error = $state('');
	let loadFailed = $state(false);
	const confirm = createConfirm();
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
		loadFailed = false;
		try {
			posts = await artPostService.listMyPosts();
		} catch (e) {
			loadFailed = true;
			error = e instanceof Error ? e.message : 'Could not load posts';
		} finally {
			loading = false;
		}
	}

	function askPublish(post: ArtPost) {
		confirm.ask({
			title: `Publish “${post.title}”?`,
			description: 'It appears on the public archive and your artist page.',
			confirmLabel: 'Publish',
			run: () => publish(post.id)
		});
	}

	function askArchive(post: ArtPost) {
		confirm.ask({
			title: `Archive “${post.title}”?`,
			description: 'It is removed from the public archive. You can publish it again later.',
			confirmLabel: 'Archive',
			destructive: true,
			run: () => archive(post.id)
		});
	}

	async function publish(id: string) {
		acting = id;
		error = '';
		try {
			const updated = await artPostService.publishMyPost(id);
			posts = posts.map((p) => (p.id === id ? updated : p));
			toast.success('Published to the archive.');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Publish failed';
		} finally {
			acting = null;
		}
	}

	async function archive(id: string) {
		acting = id;
		error = '';
		try {
			const updated = await artPostService.archiveMyPost(id);
			posts = posts.map((p) => (p.id === id ? updated : p));
			toast.success('Archived.');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Archive failed';
		} finally {
			acting = null;
		}
	}
</script>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<PageHeader eyebrow="Studio · Posts" lead="Draft, publish, and archive pieces in your file.">
		Your <em class="italic">work</em>
		{#snippet actions()}<Button href="/studio/posts/new">New draft</Button>{/snippet}
	</PageHeader>

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

	{#if error && !loadFailed}
		<Alert tone="error" class="mt-6">{error}</Alert>
	{/if}

	<div class="mt-8">
		{#if loading}
			<div class="space-y-3" role="status" aria-live="polite">
				<span class="sr-only">Loading posts</span>
				<Skeleton class="h-16 w-full" />
				<Skeleton class="h-16 w-full" />
				<Skeleton class="h-16 w-full" />
			</div>
		{:else if loadFailed}
			<ErrorState message={error} onretry={loadPosts} />
		{:else if posts.length === 0}
			<EmptyState title="Your first work starts here." body="Save a draft now and publish when you're ready. Nothing is public until you say so.">
				{#snippet action()}<Button variant="secondary" size="sm" href="/studio/posts/new">Create a draft →</Button>{/snippet}
			</EmptyState>
		{:else if filtered.length === 0}
			<EmptyState eyebrow="No matches" title="No {filter} posts." body="Try another view.">
				{#snippet action()}<Button variant="secondary" size="sm" onclick={() => (filter = 'all')}>Show all</Button>{/snippet}
			</EmptyState>
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
								<Badge tone={post.status === 'published' ? 'success' : 'neutral'}>{post.status ?? 'draft'}</Badge>
								<Button variant="link" href="/studio/posts/{post.id}">Edit</Button>
								{#if (post.status ?? 'draft') === 'draft' || post.status === 'archived'}
									<Button variant="ghost" size="sm" disabled={acting === post.id} onclick={() => askPublish(post)}>Publish</Button>
								{/if}
								{#if post.status === 'published'}
									<Button variant="ghost" size="sm" disabled={acting === post.id} onclick={() => askArchive(post)}>Archive</Button>
								{/if}
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</section>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
