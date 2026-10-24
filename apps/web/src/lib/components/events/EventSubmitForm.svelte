<script lang="ts">
	import { Alert, Button, Field, Input, Textarea, toast } from '$ui';
	import { eventsService } from '$lib/application/events';
	import type { Event as DomainEvent } from '$lib/core/domain/events';

	let { onsubmitted }: { onsubmitted?: (event: DomainEvent) => void } = $props();

	let title = $state('');
	let description = $state('');
	let eventType = $state('Opening');
	let venue = $state('');
	let city = $state('Addis Ababa');
	let startsAt = $state('');
	let sourceUrl = $state('');
	let submitting = $state(false);
	let error = $state('');
	let success = $state('');
	let fieldErrors = $state<{ title?: string; startsAt?: string }>({});

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		success = '';
		fieldErrors = {
			title: title.trim() ? undefined : 'Give the event a title.',
			startsAt: startsAt ? undefined : 'Choose when it starts.'
		};
		if (fieldErrors.title || fieldErrors.startsAt) return;
		submitting = true;
		try {
			const created = await eventsService.submit({
				title: title.trim(),
				description: description.trim() || undefined,
				event_type: eventType.trim() || undefined,
				venue: venue.trim() || undefined,
				city: city.trim() || undefined,
				source_url: sourceUrl.trim() || undefined,
				starts_at: new Date(startsAt).toISOString()
			});
			success = `“${created.title}” is with our editors. It appears on the calendar once approved.`;
			toast.success('Event submitted for review.');
			title = '';
			description = '';
			venue = '';
			sourceUrl = '';
			startsAt = '';
			onsubmitted?.(created);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Submission failed.';
		} finally {
			submitting = false;
		}
	}

</script>

<form class="max-w-xl space-y-5 rounded-sm border border-border bg-card/40 p-6 md:p-8" onsubmit={submit} novalidate>
	<div>
		<p class="t-eyebrow text-accent">Community</p>
		<h2 class="t-h3 mt-3 md:text-3xl">Submit an event</h2>
		<p class="t-body mt-2 text-muted-foreground">
			Share an opening, talk or pop-up. Submissions are reviewed by an editor before they appear on the calendar.
		</p>
	</div>

	<Field label="Title" required error={fieldErrors.title}>
		{#snippet children({ id, describedby, invalid })}
			<Input {id} aria-describedby={describedby} {invalid} bind:value={title} placeholder="Blue Hour opening" />
		{/snippet}
	</Field>

	<Field label="Starts at" required error={fieldErrors.startsAt}>
		{#snippet children({ id, describedby, invalid })}
			<Input {id} aria-describedby={describedby} {invalid} type="datetime-local" bind:value={startsAt} />
		{/snippet}
	</Field>

	<div class="grid gap-5 sm:grid-cols-2">
		<Field label="Type">
			{#snippet children({ id })}<Input {id} bind:value={eventType} placeholder="Opening" />{/snippet}
		</Field>
		<Field label="City">
			{#snippet children({ id })}<Input {id} bind:value={city} placeholder="Addis Ababa" />{/snippet}
		</Field>
	</div>

	<Field label="Venue">
		{#snippet children({ id })}<Input {id} bind:value={venue} placeholder="Addis Fine Arts" />{/snippet}
	</Field>

	<Field label="Source URL" hint="Optional. A Telegram post or event page helps us verify it.">
		{#snippet children({ id, describedby })}
			<Input {id} aria-describedby={describedby} type="url" bind:value={sourceUrl} placeholder="https://t.me/…" />
		{/snippet}
	</Field>

	<Field label="Description">
		{#snippet children({ id })}<Textarea {id} bind:value={description} rows={4} placeholder="What is it about?" />{/snippet}
	</Field>

	{#if error}<Alert tone="error">{error}</Alert>{/if}
	{#if success}<Alert tone="success" title="Submitted">{success}</Alert>{/if}

	<Button type="submit" loading={submitting}>{submitting ? 'Submitting…' : 'Submit for review'}</Button>
</form>
