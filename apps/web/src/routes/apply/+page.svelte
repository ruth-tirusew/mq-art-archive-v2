<script lang="ts">
	import { goto } from '$app/navigation';
	import { onboardingService } from '$lib/application/onboarding';
	import { ApiError } from '$lib/adapters/api/client';
	import type { ApplicantType } from '$lib/core/domain/onboarding';

	let applicantType = $state<ApplicantType>('artist');
	let displayName = $state('');
	let requestedHandle = $state('');
	let handleAvailable = $state<boolean | null>(null);
	let checkingHandle = $state(false);
	let notes = $state('');
	let submitting = $state(false);
	let error = $state('');
	let message = $state('');
	let handleTimer: ReturnType<typeof setTimeout> | undefined;

	const normalizedHandle = $derived(requestedHandle.trim().toLowerCase());
	const validHandle = $derived(/^[a-z0-9_]{3,30}$/.test(normalizedHandle));

	$effect(() => {
		const handle = normalizedHandle;
		clearTimeout(handleTimer);
		handleAvailable = null;
		checkingHandle = false;
		if (applicantType !== 'artist' || !validHandle) return;

		handleTimer = setTimeout(async () => {
			checkingHandle = true;
			try {
				const result = await onboardingService.checkHandleAvailability(handle);
				if (normalizedHandle === handle) handleAvailable = result.available;
			} catch {
				if (normalizedHandle === handle) handleAvailable = null;
			} finally {
				if (normalizedHandle === handle) checkingHandle = false;
			}
		}, 300);
		return () => clearTimeout(handleTimer);
	});

	async function submit() {
		submitting = true;
		error = '';
		message = '';
		try {
			await onboardingService.submit({
				applicant_type: applicantType,
				display_name: displayName.trim(),
				...(applicantType === 'artist' ? { requested_handle: normalizedHandle } : {}),
				notes: notes.trim()
			});
			message = 'Application submitted. We will review it shortly.';
			await goto('/apply/status');
		} catch (e) {
			if (e instanceof ApiError && e.status === 409) {
				error = 'You already have a pending or approved application.';
			} else {
				error = e instanceof Error ? e.message : 'Could not submit application';
			}
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Apply — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-2xl px-6 py-24 md:px-10">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Onboarding</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Apply to Artiv</h1>
	<p class="mt-4 text-sm leading-relaxed text-muted-foreground">
		Submit your name and supporting links. After review, approved artists get access to the studio.
	</p>

	<form class="mt-10 space-y-5" onsubmit={(e) => { e.preventDefault(); void submit(); }}>
		<fieldset>
			<legend class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				Applicant type
			</legend>
			<div class="grid grid-cols-2 gap-2">
				{#each ['artist', 'institution'] as type}
					<button
						type="button"
						class="rounded-sm border px-4 py-3 font-mono text-[11px] uppercase tracking-[0.18em] transition {applicantType === type ? 'border-foreground bg-foreground text-background' : 'border-border text-muted-foreground hover:text-foreground'}"
						onclick={() => applicantType = type as ApplicantType}
					>
						{type}
					</button>
				{/each}
			</div>
		</fieldset>
		<div>
			<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="display_name">
				Display name
			</label>
			<input id="display_name" class="field" bind:value={displayName} required />
		</div>
		{#if applicantType === 'artist'}
			<div>
				<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="requested_handle">
					Requested handle
				</label>
				<div class="relative">
					<span class="pointer-events-none absolute left-4 top-3 text-sm text-muted-foreground">@</span>
					<input
						id="requested_handle"
						class="field pl-8"
						bind:value={requestedHandle}
						pattern={'[a-z0-9_]{3,30}'}
						minlength="3"
						maxlength="30"
						required
						autocomplete="off"
						aria-describedby="handle-status"
					/>
				</div>
				<p id="handle-status" class="mt-2 text-xs {handleAvailable === false || (normalizedHandle && !validHandle) ? 'text-destructive' : 'text-muted-foreground'}">
					{#if normalizedHandle && !validHandle}
						Use 3–30 lowercase letters, numbers, or underscores.
					{:else if checkingHandle}
						Checking availability…
					{:else if handleAvailable === true}
						@{normalizedHandle} is available.
					{:else if handleAvailable === false}
						@{normalizedHandle} is already taken.
					{:else}
						This becomes your shareable profile address.
					{/if}
				</p>
			</div>
		{/if}
		<div>
			<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="notes">
				Portfolio / links
			</label>
			<textarea
				id="notes"
				class="field min-h-32"
				placeholder="Instagram, Telegram, website, or anything that helps us verify your practice."
				bind:value={notes}
			></textarea>
		</div>

		{#if error}
			<p class="text-sm text-destructive" role="alert">{error}</p>
		{/if}
		{#if message}
			<p class="text-sm text-accent">{message}</p>
		{/if}

		<button
			type="submit"
			class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90 disabled:opacity-50"
			disabled={submitting || !displayName.trim() || (applicantType === 'artist' && (!validHandle || handleAvailable !== true))}
		>
			{submitting ? 'Submitting…' : 'Submit application'}
		</button>
	</form>

	<p class="mt-8 text-sm text-muted-foreground">
		Already applied?
		<a href="/apply/status" class="text-accent underline underline-offset-4">Check status</a>
	</p>
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
