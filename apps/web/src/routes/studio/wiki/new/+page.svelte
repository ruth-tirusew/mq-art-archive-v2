<script lang="ts">
	import { goto } from '$app/navigation';
	import { wikiSubmissionsService } from '$lib/application/wikiSubmissions';
	import { Alert, Button, Field, Input, PageHeader, Textarea, toast } from '$ui';

	let title = $state('');
	let body = $state('');
	let submitting = $state(false);
	let error = $state('');
	let fieldErrors = $state<{ title?: string; body?: string }>({});

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		fieldErrors = {
			title: title.trim() ? undefined : 'Give the article a title.',
			body: body.trim() ? undefined : 'Write something before submitting.'
		};
		if (fieldErrors.title || fieldErrors.body) return;
		submitting = true;
		try {
			await wikiSubmissionsService.submit({ title: title.trim(), body: body.trim() });
			toast.success('Submitted. An editor will review it before it is published.');
			await goto('/studio/wiki');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not submit article';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head><title>Submit to the wiki — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<div class="mb-8"><Button variant="link" href="/studio/wiki">← Submissions</Button></div>
	<PageHeader eyebrow="Community wiki" lead="Editors review your contribution before it is published.">
		Submit an <em class="italic">article</em>
	</PageHeader>

	<form class="mt-10 space-y-6" onsubmit={submit} novalidate>
		<Field label="Title" required error={fieldErrors.title}>
			{#snippet children({ id, describedby, invalid })}
				<Input {id} aria-describedby={describedby} {invalid} bind:value={title} />
			{/snippet}
		</Field>
		<Field label="Article body" required error={fieldErrors.body} hint="Separate paragraphs with a blank line.">
			{#snippet children({ id, describedby, invalid })}
				<Textarea {id} aria-describedby={describedby} {invalid} rows={14} bind:value={body} />
			{/snippet}
		</Field>
		{#if error}<Alert tone="error">{error}</Alert>{/if}
		<Button type="submit" loading={submitting}>{submitting ? 'Submitting…' : 'Submit for review'}</Button>
	</form>
</section>
