<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { artPostService } from '$lib/application/artPosts';
	import MediaUploader from '$lib/components/MediaUploader.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import { Alert, Badge, Button, ConfirmDialog, createConfirm, ErrorState, Field, Input, PageHeader, Textarea, toast } from '$ui';

	const id = $derived(($page.params.id ?? '') as string);
	const isNew = $derived(id === 'new');

	let loaded = $state(false);
	let saving = $state(false);
	let acting = $state(false);
	let error = $state('');
	let fieldErrors = $state<{ title?: string; year?: string }>({});
	const confirm = createConfirm();
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

	async function load() {
		error = '';
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
	}

	onMount(load);

	function requestSave() {
		const yearOk = !year.trim() || (Number.isInteger(Number(year)) && Number(year) > 0);
		fieldErrors = {
			title: title.trim() ? undefined : 'A title is required.',
			year: yearOk ? undefined : 'Enter a four-digit year, or leave it blank.'
		};
		if (fieldErrors.title || fieldErrors.year) return;
		void save();
	}

	async function save() {
		saving = true;
		error = '';
		try {
			const input = formInput();
			if (isNew) {
				const created = await artPostService.createDraft(input);
				await goto(`/studio/posts/${created.id}`, { replaceState: true });
				apply(created);
				toast.success('Draft created.');
			} else {
				apply(await artPostService.updateMyPost(id, input));
				toast.success('Saved.');
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
			toast.success('Published to the archive.');
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
			toast.success('Unpublished. It is a draft again.');
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
			toast.success('Archived.');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Archive failed';
		} finally {
			acting = false;
		}
	}

	function ask(kind: 'publish' | 'unpublish' | 'archive' | 'delete') {
		const copy = {
			publish: { title: 'Publish this work?', description: 'It appears on the public archive and your artist page.', confirmLabel: 'Publish', destructive: false, run: publish },
			unpublish: { title: 'Unpublish this work?', description: 'It is removed from the public archive and becomes a draft.', confirmLabel: 'Unpublish', destructive: true, run: unpublish },
			archive: { title: 'Archive this work?', description: 'It is removed from the public archive. You can publish it again later.', confirmLabel: 'Archive', destructive: true, run: archive },
			delete: { title: 'Delete this post?', description: 'This cannot be undone.', confirmLabel: 'Delete', destructive: true, run: remove }
		};
		confirm.ask(copy[kind]);
	}

	async function remove() {
		acting = true;
		error = '';
		try {
			await artPostService.deleteMyPost(id);
			toast.success('Post deleted.');
			await goto('/studio/posts');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Delete failed';
			acting = false;
		}
	}
</script>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<PageHeader eyebrow="Studio · Posts">
		{isNew ? 'New' : 'Edit'} <em class="italic">{isNew ? 'draft' : 'post'}</em>
		{#snippet actions()}
			{#if post}<Badge tone={post.status === 'published' ? 'success' : 'neutral'}>{post.status ?? 'draft'}</Badge>{/if}
		{/snippet}
	</PageHeader>

	{#if !loaded && !error}
		<div class="mt-8 space-y-4" role="status" aria-live="polite">
			<span class="sr-only">Loading post</span>
			<Skeleton class="h-12 w-full" />
			<Skeleton class="h-48 w-full" />
		</div>
	{:else if error && !loaded}
		<div class="mt-8 space-y-4">
			<ErrorState message={error} onretry={load} />
			<Button variant="link" href="/studio/posts">← Back to posts</Button>
		</div>
	{:else}
		<form class="mt-10 space-y-5" onsubmit={(e) => { e.preventDefault(); requestSave(); }} novalidate>
			<Field label="Title" required error={fieldErrors.title}>
				{#snippet children({ id, describedby, invalid })}
					<Input {id} aria-describedby={describedby} {invalid} bind:value={title} />
				{/snippet}
			</Field>
			<Field label="Description">
				{#snippet children({ id })}<Textarea {id} rows={4} bind:value={description} />{/snippet}
			</Field>
			<div class="grid gap-5 sm:grid-cols-2">
				<Field label="Medium">{#snippet children({ id })}<Input {id} bind:value={medium} placeholder="Oil on canvas" />{/snippet}</Field>
				<Field label="Year" error={fieldErrors.year}>
					{#snippet children({ id, describedby, invalid })}
						<Input {id} aria-describedby={describedby} {invalid} bind:value={year} inputmode="numeric" placeholder="2024" />
					{/snippet}
				</Field>
			</div>
			<div class="grid gap-5 sm:grid-cols-2">
				<Field label="Dimensions">{#snippet children({ id })}<Input {id} bind:value={dimensions} placeholder="90 × 120 cm" />{/snippet}</Field>
				<Field label="City">{#snippet children({ id })}<Input {id} bind:value={city} />{/snippet}</Field>
			</div>
			<Field label="Style">{#snippet children({ id })}<Input {id} bind:value={style} />{/snippet}</Field>
			<Field label="Palette" hint="Comma-separated hex colours, e.g. #6b2e2b, #c4a882">
				{#snippet children({ id, describedby })}<Input {id} aria-describedby={describedby} bind:value={paletteText} />{/snippet}
			</Field>
			<div class="space-y-3 rounded-sm border border-border/60 p-4">
				<p class="t-eyebrow text-foreground">Images</p>
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
									<Button
										variant="ghost"
										size="sm"
										class="mt-2 !px-0 text-destructive"
										onclick={() => (mediaUrls = mediaUrls.filter((_, itemIndex) => itemIndex !== index))}
									>
										Remove
									</Button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</div>

			{#if error}<Alert tone="error">{error}</Alert>{/if}

			<div class="flex flex-wrap items-center gap-3 pt-2">
				<Button type="submit" loading={saving}>{saving ? 'Saving…' : isNew ? 'Create draft' : 'Save'}</Button>
				{#if !isNew && (post?.status === 'draft' || post?.status === 'archived')}
					<Button variant="accent" disabled={acting} onclick={() => ask('publish')}>Publish</Button>
				{/if}
				{#if !isNew && post?.status === 'published'}
					<Button variant="secondary" disabled={acting} onclick={() => ask('unpublish')}>Unpublish</Button>
					<Button variant="secondary" disabled={acting} onclick={() => ask('archive')}>Archive</Button>
				{/if}
				{#if !isNew && post?.status !== 'published'}
					<Button variant="ghost" class="text-destructive" disabled={acting} onclick={() => ask('delete')}>Delete</Button>
				{/if}
				<Button variant="link" href="/studio/posts">Back</Button>
			</div>
		</form>
	{/if}
</section>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
