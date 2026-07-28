<script lang="ts">
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

	async function submit(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		success = '';
		if (!title.trim() || !startsAt) {
			error = 'Title and start date are required.';
			return;
		}
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
			success = `Submitted “${created.title}” for review.`;
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

	const fieldClass =
		'mt-1.5 w-full rounded-sm border border-border bg-background px-3 py-2.5 text-sm text-foreground placeholder:text-muted-foreground/60 focus:border-foreground focus:outline-none';
	const labelClass = 'font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground';
</script>

<form
	class="max-w-xl space-y-5 rounded-sm border border-border bg-card/40 p-6 md:p-8"
	onsubmit={submit}
>
	<div>
		<h2 class="font-display text-2xl text-foreground md:text-3xl">Submit an event</h2>
		<p class="mt-2 text-sm text-muted-foreground">
			Share an opening, talk or pop-up. Submissions appear as pending until an editor reviews them.
		</p>
	</div>

	<label class="block">
		<span class={labelClass}>Title</span>
		<input class={fieldClass} bind:value={title} required placeholder="Blue Hour opening" />
	</label>

	<label class="block">
		<span class={labelClass}>Starts at</span>
		<input class={fieldClass} type="datetime-local" bind:value={startsAt} required />
	</label>

	<div class="grid gap-5 sm:grid-cols-2">
		<label class="block">
			<span class={labelClass}>Type</span>
			<input class={fieldClass} bind:value={eventType} placeholder="Opening" />
		</label>
		<label class="block">
			<span class={labelClass}>City</span>
			<input class={fieldClass} bind:value={city} placeholder="Addis Ababa" />
		</label>
	</div>

	<label class="block">
		<span class={labelClass}>Venue</span>
		<input class={fieldClass} bind:value={venue} placeholder="Addis Fine Arts" />
	</label>

	<label class="block">
		<span class={labelClass}>Source URL (optional)</span>
		<input class={fieldClass} bind:value={sourceUrl} placeholder="https://t.me/…" />
	</label>

	<label class="block">
		<span class={labelClass}>Description</span>
		<textarea class="{fieldClass} min-h-28" bind:value={description} placeholder="What is it about?"
		></textarea>
	</label>

	{#if error}
		<p class="text-sm text-destructive">{error}</p>
	{/if}
	{#if success}
		<p class="text-sm text-foreground">{success}</p>
	{/if}

	<button
		type="submit"
		disabled={submitting}
		class="rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background transition hover:bg-accent disabled:opacity-60"
	>
		{submitting ? 'Submitting…' : 'Submit for review'}
	</button>
</form>
