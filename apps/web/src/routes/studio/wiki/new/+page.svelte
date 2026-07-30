<script lang="ts">
	import { goto } from '$app/navigation';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';

	let title = $state('');
	let body = $state('');
	let submitting = $state(false);
	let error = $state('');

	async function submit() {
		submitting = true;
		error = '';
		try {
			await wikiSubmissionsService.submit({ title: title.trim(), body: body.trim() });
			await goto('/studio/wiki');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not submit article';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>Submit to the wiki — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<a href="/studio/wiki" class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">← Submissions</a>
	<p class="mt-8 font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Community wiki</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Submit an article</h1>
	<p class="mt-3 text-sm text-muted-foreground">Editors will review your contribution before it is published.</p>

	<form class="mt-8 space-y-5" onsubmit={(e) => { e.preventDefault(); void submit(); }}>
		<div>
			<label for="title" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Title</label>
			<input id="title" class="field" bind:value={title} required />
		</div>
		<div>
			<label for="body" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Article body</label>
			<textarea id="body" class="field min-h-80" bind:value={body} required></textarea>
		</div>
		{#if error}<p class="text-sm text-destructive" role="alert">{error}</p>{/if}
		<button type="submit" disabled={submitting || !title.trim() || !body.trim()} class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50">
			{submitting ? 'Submitting…' : 'Submit for review'}
		</button>
	</form>
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
