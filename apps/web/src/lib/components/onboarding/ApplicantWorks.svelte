<script lang="ts">
	import { onMount } from 'svelte';
	import BlurhashImage from '$lib/components/BlurhashImage.svelte';
	import MediaUploader from '$lib/components/MediaUploader.svelte';
	import { applicantWorksService } from '$lib/composition';
	import type { ApplicantWork } from '$lib/core/domain/onboarding';
	import { MIN_APPLICATION_WORKS, worksShortfall } from '$lib/core/domain/onboarding';
	import type { UploadedMedia } from '$lib/core/domain/media';
	import { Alert, Button, Field, Input } from '$ui';

	type Props = {
		/** Called whenever the list changes, so the page can decide whether to let the applicant continue. */
		onchange?: (works: ApplicantWork[]) => void;
	};

	let { onchange }: Props = $props();

	let works = $state<ApplicantWork[]>([]);
	let loading = $state(true);
	let loadFailed = $state(false);
	let error = $state('');
	let busy = $state(false);

	// The image just uploaded, waiting for a title before it becomes a work.
	let pending = $state<UploadedMedia | null>(null);
	let title = $state('');
	let medium = $state('');
	let year = $state('');
	let titleError = $state('');

	const shortfall = $derived(worksShortfall(works.length));
	const atLimit = $derived(works.length >= applicantWorksService.maxWorks);

	function set(next: ApplicantWork[]) {
		works = next;
		onchange?.(next);
	}

	async function load() {
		loading = true;
		loadFailed = false;
		try {
			set(await applicantWorksService.list());
		} catch {
			loadFailed = true;
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function uploaded(media: UploadedMedia) {
		error = '';
		pending = media;
		title = medium = year = '';
		titleError = '';
	}

	async function add(event?: Event) {
		event?.preventDefault();
		if (!pending?.id) return;
		if (!title.trim()) {
			titleError = 'Give this work a title.';
			return;
		}
		const parsedYear = year.trim() ? Number(year) : null;
		if (parsedYear !== null && !Number.isInteger(parsedYear)) {
			titleError = '';
			error = 'Year should be a number, like 2024.';
			return;
		}
		busy = true;
		error = '';
		try {
			const work = await applicantWorksService.add({ media_asset_id: pending.id, title, medium, year: parsedYear });
			set([...works, work]);
			pending = null;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not add this work.';
		} finally {
			busy = false;
		}
	}

	// Enter adds the work (and must not submit the whole application, which this sits inside).
	function enterAdds(event: KeyboardEvent) {
		if (event.key !== 'Enter') return;
		event.preventDefault();
		void add();
	}

	async function remove(work: ApplicantWork) {
		busy = true;
		error = '';
		try {
			await applicantWorksService.remove(work.id);
			set(works.filter((w) => w.id !== work.id));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not remove this work.';
		} finally {
			busy = false;
		}
	}
</script>

<div class="space-y-6" data-testid="applicant-works">
	<div>
		<p class="t-eyebrow text-accent">Your art</p>
		<h2 class="t-h3 mt-2">Show us at least {MIN_APPLICATION_WORKS} pieces of your work</h2>
		<p class="mt-2 text-sm text-muted-foreground">
			Upload your best pieces, one at a time, and give each a title. These are what our team looks at first, and once you're approved
			they're a head start on your profile.
		</p>
	</div>

	<p class="font-mono text-xs uppercase tracking-[0.18em] {shortfall === 0 ? 'text-success' : 'text-muted-foreground'}" role="status" aria-live="polite">
		{#if shortfall === 0}
			{works.length} of {MIN_APPLICATION_WORKS} · you have enough to apply
		{:else}
			{works.length} of {MIN_APPLICATION_WORKS} · add {shortfall} more
		{/if}
	</p>

	{#if loading}
		<p class="text-sm text-muted-foreground">Loading your works…</p>
	{:else if loadFailed}
		<Alert tone="error">
			We couldn't load your works.
			<button type="button" class="underline underline-offset-4" onclick={load}>Try again</button>
		</Alert>
	{:else}
		{#if works.length > 0}
			<ul class="grid grid-cols-2 gap-4 sm:grid-cols-3">
				{#each works as work (work.id)}
					<li class="group" data-testid="applicant-work">
						<div class="aspect-[4/5] overflow-hidden rounded-sm bg-muted">
							<BlurhashImage src={work.url} alt={work.title} blurhash={work.blurhash} width={320} class="h-full w-full object-cover" />
						</div>
						<div class="mt-2 flex items-start justify-between gap-2">
							<div class="min-w-0">
								<p class="truncate text-sm text-foreground">{work.title}</p>
								<p class="truncate font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
									{[work.medium, work.year].filter(Boolean).join(' · ')}
								</p>
							</div>
							<button
								type="button"
								class="shrink-0 font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground underline underline-offset-4 hover:text-destructive disabled:opacity-50"
								disabled={busy}
								onclick={() => remove(work)}
								aria-label="Remove {work.title}"
							>
								Remove
							</button>
						</div>
					</li>
				{/each}
			</ul>
		{/if}

		{#if pending}
			<!-- Not a <form>: this sits inside the application's own form, and forms can't nest. -->
			<div class="space-y-4 rounded-sm border border-border/70 p-4" role="group" aria-label="Add this work">
				<div class="flex gap-4">
					<div class="h-24 w-20 shrink-0 overflow-hidden rounded-sm bg-muted">
						<BlurhashImage src={pending.secure_url} alt="Uploaded work" blurhash={pending.blurhash} width={160} class="h-full w-full object-cover" />
					</div>
					<div class="min-w-0 flex-1 space-y-3">
						<Field label="Title" required error={titleError || undefined}>
							{#snippet children({ id, describedby, invalid })}
								<Input {id} aria-describedby={describedby} {invalid} bind:value={title} maxlength={200} onkeydown={enterAdds} />
							{/snippet}
						</Field>
						<div class="grid grid-cols-2 gap-3">
							<Field label="Medium" hint="Optional">
								{#snippet children({ id, describedby })}
									<Input {id} aria-describedby={describedby} bind:value={medium} maxlength={100} onkeydown={enterAdds} />
								{/snippet}
							</Field>
							<Field label="Year" hint="Optional">
								{#snippet children({ id, describedby })}
									<Input {id} inputmode="numeric" aria-describedby={describedby} bind:value={year} maxlength={4} onkeydown={enterAdds} />
								{/snippet}
							</Field>
						</div>
					</div>
				</div>
				<div class="flex gap-3">
					<Button type="button" disabled={busy} onclick={add}>{busy ? 'Adding…' : 'Add this work'}</Button>
					<Button type="button" variant="ghost" disabled={busy} onclick={() => (pending = null)}>Discard</Button>
				</div>
			</div>
		{:else if atLimit}
			<p class="text-sm text-muted-foreground">That's the most we can take ({applicantWorksService.maxWorks}). Remove one to add another.</p>
		{:else}
			<MediaUploader scope="applicant" label={works.length === 0 ? 'Upload your first piece' : 'Upload another piece'} showPreview={false} onUploaded={uploaded} />
		{/if}
	{/if}

	{#if error}<Alert tone="error">{error}</Alert>{/if}
</div>
