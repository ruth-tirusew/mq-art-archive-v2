<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { artPostService } from '$lib/application/artPosts';
	import MediaUploader from '$lib/components/MediaUploader.svelte';
	import type { ArtPost } from '$lib/core/domain/art';

	const id = $derived(($page.params.id ?? '') as string);
	const isNew = $derived(id === 'new');

	let loaded = $state(false);
	let saving = $state(false);
	let acting = $state(false);
	let error = $state('');
	let message = $state('');
	let post = $state<ArtPost | null>(null);

	let title = $state('');
	let description = $state('');
	let medium = $state('');
	let year = $state('');
	let dimensions = $state('');
	let city = $state('');
	let style = $state('');
	let paletteText = $state('');
	let mediaUrls = $state<string[]>([]);

	function apply(p: ArtPost) {
		post = p;
		title = p.title;
		description = p.description ?? '';
		medium = p.medium ?? '';
		year = p.year != null ? String(p.year) : '';
		dimensions = p.dimensions ?? '';
		city = p.city ?? '';
		style = p.style ?? '';
		paletteText = (p.palette ?? []).join(', ');
		mediaUrls = (p.media ?? []).map((m) => m.url);
	}

	function formInput() {
		const yearNum = year.trim() ? Number(year) : null;
		return {
			title: title.trim(),
			description,
			medium,
			year: yearNum != null && Number.isFinite(yearNum) ? yearNum : null,
			dimensions,
			city,
			style,
			palette: paletteText
				.split(',')
				.map((s) => s.trim())
				.filter(Boolean),
			media_urls: mediaUrls
		};
	}

	onMount(async () => {
		if (isNew) {
			loaded = true;
			return;
		}
		try {
			apply(await artPostService.getMyPost(id));
			loaded = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load post';
		}
	});

	async function save() {
		saving = true;
		message = '';
		error = '';
		try {
			const input = formInput();
			if (isNew) {
				const created = await artPostService.createDraft(input);
				await goto(`/studio/posts/${created.id}`, { replaceState: true });
				apply(created);
				message = 'Draft created.';
			} else {
				apply(await artPostService.updateMyPost(id, input));
				message = 'Saved.';
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Save failed';
		} finally {
			saving = false;
		}
	}

	async function publish() {
		acting = true;
		error = '';
		try {
			apply(await artPostService.publishMyPost(id));
			message = 'Published.';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Publish failed';
		} finally {
			acting = false;
		}
	}

	async function unpublish() {
		acting = true;
		error = '';
		try {
			apply(await artPostService.unpublishMyPost(id));
			message = 'Unpublished to draft.';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Unpublish failed';
		} finally {
			acting = false;
		}
	}

	async function archive() {
		acting = true;
		error = '';
		try {
			apply(await artPostService.archiveMyPost(id));
			message = 'Archived.';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Archive failed';
		} finally {
			acting = false;
		}
	}

	async function remove() {
		if (!confirm('Delete this post? Published works must be archived first.')) return;
		acting = true;
		error = '';
		try {
			await artPostService.deleteMyPost(id);
			await goto('/studio/posts');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
			acting = false;
		}
	}
</script>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Studio · Posts</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">{isNew ? 'New draft' : 'Edit post'}</h1>
	{#if post}
		<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
			Status · {post.status ?? 'draft'}
		</p>
	{/if}

	{#if !loaded && !error}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error && !loaded}
		<p class="mt-6 text-sm text-destructive" role="alert">{error}</p>
	{:else}
		<form
			class="mt-10 space-y-4"
			onsubmit={(e) => {
				e.preventDefault();
				void save();
			}}
		>
			<input class="field" placeholder="Title" bind:value={title} required />
			<textarea class="field min-h-24" placeholder="Description" bind:value={description}></textarea>
			<div class="grid gap-4 sm:grid-cols-2">
				<input class="field" placeholder="Medium" bind:value={medium} />
				<input class="field" placeholder="Year" bind:value={year} inputmode="numeric" />
			</div>
			<div class="grid gap-4 sm:grid-cols-2">
				<input class="field" placeholder="Dimensions" bind:value={dimensions} />
				<input class="field" placeholder="City" bind:value={city} />
			</div>
			<input class="field" placeholder="Style" bind:value={style} />
			<input class="field" placeholder="Palette (comma-separated hex)" bind:value={paletteText} />
			<div class="space-y-3 rounded-sm border border-border/60 p-4">
				<MediaUploader
					onUploaded={(media) => {
						if (!mediaUrls.includes(media.secure_url)) mediaUrls = [...mediaUrls, media.secure_url];
					}}
				/>
				{#if mediaUrls.length}
					<ul class="grid gap-3 sm:grid-cols-2">
						{#each mediaUrls as url, index}
							<li class="flex gap-3 border border-border/60 p-2">
								<img src={url} alt="" class="h-16 w-16 shrink-0 object-cover" />
								<div class="min-w-0 flex-1">
									<p class="truncate text-xs text-muted-foreground">{url}</p>
									<button
										type="button"
										class="mt-2 font-mono text-[10px] uppercase tracking-[0.15em] text-destructive"
										onclick={() => (mediaUrls = mediaUrls.filter((_, itemIndex) => itemIndex !== index))}
									>
										Remove
									</button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			{#if message}
				<p class="text-sm text-accent">{message}</p>
			{/if}
			{#if error}
				<p class="text-sm text-destructive" role="alert">{error}</p>
			{/if}

			<div class="flex flex-wrap gap-3 pt-2">
				<button
					type="submit"
					class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
					disabled={saving || !title.trim()}
				>
					{saving ? 'Saving…' : isNew ? 'Create draft' : 'Save'}
				</button>
				{#if !isNew && (post?.status === 'draft' || post?.status === 'archived')}
					<button
						type="button"
						class="rounded-sm border border-accent px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-accent disabled:opacity-50"
						disabled={acting}
						onclick={() => void publish()}
					>
						Publish
					</button>
				{/if}
				{#if !isNew && post?.status === 'published'}
					<button
						type="button"
						class="rounded-sm border border-border px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-foreground disabled:opacity-50"
						disabled={acting}
						onclick={() => void unpublish()}
					>
						Unpublish
					</button>
					<button
						type="button"
						class="rounded-sm border border-border px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground disabled:opacity-50"
						disabled={acting}
						onclick={() => void archive()}
					>
						Archive
					</button>
				{/if}
				{#if !isNew && post?.status !== 'published'}
					<button
						type="button"
						class="rounded-sm border border-destructive/40 px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-destructive disabled:opacity-50"
						disabled={acting}
						onclick={() => void remove()}
					>
						Delete
					</button>
				{/if}
				<a
					href="/studio/posts"
					class="px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground hover:text-foreground"
				>
					Back
				</a>
			</div>
		</form>
	{/if}
</section>

<style>
	.field {
		width: 100%;
		border: 1px solid color-mix(in oklab, var(--border) 70%, transparent);
		background: color-mix(in oklab, var(--card) 30%, transparent);
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		color: var(--foreground);
	}
	.field:focus {
		outline: 2px solid color-mix(in oklab, var(--accent) 50%, transparent);
		outline-offset: 2px;
	}
</style>
