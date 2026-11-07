<script lang="ts">
	import { onMount } from 'svelte';
	import BlurhashImage from '$lib/components/BlurhashImage.svelte';
	import { validateImageFile } from '$lib/application/media';
	import { applicantWorksService } from '$lib/composition';
	import type { ApplicantWork, WorkDraft } from '$lib/core/domain/onboarding';
	import { MIN_APPLICATION_WORKS, worksShortfall, workDraftProblem } from '$lib/core/domain/onboarding';
	import { Alert, Field, Input } from '$ui';

	type Props = {
		/** Pieces chosen but not uploaded yet. Nothing here touches the network until the application is submitted. */
		drafts: WorkDraft[];
		/** Works already stored for this applicant. Only someone re-applying after a rejection has any. */
		existing: ApplicantWork[];
		/** Fetch the stored works. False for a first-time applicant, who can't have any. */
		loadExisting?: boolean;
		/** Show what's wrong with each piece (after they try to continue). */
		showProblems?: boolean;
	};

	let { drafts = $bindable(), existing = $bindable(), loadExisting = false, showProblems = false }: Props = $props();

	let error = $state('');
	let loadFailed = $state(false);
	let busyId = $state('');

	const total = $derived(existing.length + drafts.length);
	const shortfall = $derived(worksShortfall(total));
	const atLimit = $derived(total >= applicantWorksService.maxWorks);

	async function loadStored() {
		loadFailed = false;
		try {
			existing = await applicantWorksService.list();
		} catch {
			loadFailed = true;
		}
	}

	onMount(() => {
		if (loadExisting) void loadStored();
	});

	function choose(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const files = Array.from(input.files ?? []);
		input.value = '';
		const problems: string[] = [];
		const added: WorkDraft[] = [];
		for (const file of files) {
			if (total + added.length >= applicantWorksService.maxWorks) {
				problems.push(`You can add up to ${applicantWorksService.maxWorks} pieces.`);
				break;
			}
			const problem = validateImageFile(file);
			if (problem) {
				problems.push(`${file.name}: ${problem}`);
				continue;
			}
			added.push({
				key: crypto.randomUUID(),
				file,
				previewUrl: URL.createObjectURL(file),
				// A sensible starting title from the file name, which the applicant can change.
				title: file.name.replace(/\.[^.]+$/, '').replace(/[_-]+/g, ' ').trim(),
				medium: '',
				year: '',
				status: 'pending'
			});
		}
		error = problems.join(' ');
		drafts = [...drafts, ...added];
	}

	async function removeDraft(draft: WorkDraft) {
		error = '';
		// One that was already uploaded (a retry after a failed submit) also has to leave the server.
		if (draft.workId) {
			busyId = draft.key;
			try {
				await applicantWorksService.remove(draft.workId);
			} catch (e) {
				error = e instanceof Error ? e.message : 'Could not remove this piece.';
				busyId = '';
				return;
			}
			busyId = '';
		}
		URL.revokeObjectURL(draft.previewUrl);
		drafts = drafts.filter((d) => d.key !== draft.key);
	}

	async function removeStored(work: ApplicantWork) {
		error = '';
		busyId = work.id;
		try {
			await applicantWorksService.remove(work.id);
			existing = existing.filter((w) => w.id !== work.id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not remove this piece.';
		} finally {
			busyId = '';
		}
	}
</script>

<div class="space-y-6" data-testid="applicant-works">
	<div>
		<p class="t-eyebrow text-accent">Your art</p>
		<h2 class="t-h3 mt-2">Show us at least {MIN_APPLICATION_WORKS} pieces of your work</h2>
		<p class="mt-2 text-sm text-muted-foreground">
			Choose your best pieces and give each a title. Nothing is uploaded yet: you'll see everything on the next step, and your images are only
			sent when you submit your application.
		</p>
	</div>

	<p class="font-mono text-xs uppercase tracking-[0.18em] {shortfall === 0 ? 'text-success' : 'text-muted-foreground'}" role="status" aria-live="polite">
		{#if shortfall === 0}
			{total} of {MIN_APPLICATION_WORKS} · you have enough to apply
		{:else}
			{total} of {MIN_APPLICATION_WORKS} · add {shortfall} more
		{/if}
	</p>

	{#if loadFailed}
		<Alert tone="error">
			We couldn't load the works from your earlier application.
			<button type="button" class="underline underline-offset-4" onclick={loadStored}>Try again</button>
		</Alert>
	{/if}

	{#if existing.length > 0}
		<div>
			<p class="mb-3 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">From your earlier application</p>
			<ul class="grid grid-cols-2 gap-4 sm:grid-cols-3">
				{#each existing as work (work.id)}
					<li data-testid="stored-work">
						<div class="aspect-[4/5] overflow-hidden rounded-sm bg-muted">
							<BlurhashImage src={work.url} alt={work.title} blurhash={work.blurhash} width={320} class="h-full w-full object-cover" />
						</div>
						<div class="mt-2 flex items-start justify-between gap-2">
							<p class="min-w-0 truncate text-sm text-foreground">{work.title}</p>
							<button
								type="button"
								class="shrink-0 font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground underline underline-offset-4 hover:text-destructive disabled:opacity-50"
								disabled={busyId === work.id}
								onclick={() => removeStored(work)}
								aria-label="Remove {work.title}"
							>
								Remove
							</button>
						</div>
					</li>
				{/each}
			</ul>
		</div>
	{/if}

	{#if drafts.length > 0}
		<ul class="space-y-4" data-testid="work-drafts">
			{#each drafts as draft (draft.key)}
				{@const problem = showProblems ? workDraftProblem(draft) : null}
				<li class="flex gap-4 rounded-sm border border-border/70 p-4" data-testid="work-draft">
					<div class="h-28 w-24 shrink-0 overflow-hidden rounded-sm bg-muted">
						<!-- A local preview of the file the applicant chose; it has not been uploaded. -->
						<img src={draft.previewUrl} alt="Preview of {draft.title || draft.file.name}" class="h-full w-full object-cover" />
					</div>
					<div class="min-w-0 flex-1 space-y-3">
						<Field label="Title" required error={problem ?? undefined}>
							{#snippet children({ id, describedby, invalid })}
								<Input {id} aria-describedby={describedby} {invalid} bind:value={draft.title} maxlength={200} disabled={draft.status === 'done'} />
							{/snippet}
						</Field>
						<div class="grid grid-cols-2 gap-3">
							<Field label="Medium" hint="Optional">
								{#snippet children({ id, describedby })}
									<Input {id} aria-describedby={describedby} bind:value={draft.medium} maxlength={100} disabled={draft.status === 'done'} />
								{/snippet}
							</Field>
							<Field label="Year" hint="Optional">
								{#snippet children({ id, describedby })}
									<Input {id} inputmode="numeric" aria-describedby={describedby} bind:value={draft.year} maxlength={4} disabled={draft.status === 'done'} />
								{/snippet}
							</Field>
						</div>
						<div class="flex items-center justify-between gap-3">
							<p class="font-mono text-[10px] uppercase tracking-[0.15em] {draft.status === 'failed' ? 'text-destructive' : 'text-muted-foreground'}" role="status">
								{#if draft.status === 'uploading'}Uploading…
								{:else if draft.status === 'done'}Uploaded
								{:else if draft.status === 'failed'}Upload failed{draft.error ? `: ${draft.error}` : ''}
								{:else}Not uploaded yet{/if}
							</p>
							<button
								type="button"
								class="shrink-0 font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground underline underline-offset-4 hover:text-destructive disabled:opacity-50"
								disabled={busyId === draft.key || draft.status === 'uploading'}
								onclick={() => removeDraft(draft)}
								aria-label="Remove {draft.title || draft.file.name}"
							>
								Remove
							</button>
						</div>
					</div>
				</li>
			{/each}
		</ul>
	{/if}

	{#if atLimit}
		<p class="text-sm text-muted-foreground">That's the most we can take ({applicantWorksService.maxWorks}). Remove one to add another.</p>
	{:else}
		<label class="block">
			<span class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				{total === 0 ? 'Choose your pieces' : 'Choose more pieces'}
			</span>
			<input
				type="file"
				multiple
				accept="image/jpeg,image/png,image/webp"
				onchange={choose}
				class="block w-full text-sm text-muted-foreground file:mr-4 file:rounded-sm file:border file:border-border file:bg-transparent file:px-4 file:py-2 file:text-foreground"
			/>
			<span class="mt-1 block text-xs text-muted-foreground">JPEG, PNG or WebP, up to 10 MB each. You can pick several at once.</span>
		</label>
	{/if}

	{#if error}<Alert tone="error">{error}</Alert>{/if}
</div>
