<script lang="ts">
	import { validateImageFile } from '$lib/application/media';
	import BlurhashImage from '$lib/components/BlurhashImage.svelte';
	import type { ApplicantPortrait, PortraitDraft } from '$lib/core/domain/onboarding';
	import { Alert } from '$ui';

	type Props = {
		/** The photo just chosen, not uploaded yet. Nothing touches the network until the application is submitted. */
		draft: PortraitDraft | null;
		/** A portrait already stored for this applicant (someone re-applying after a rejection). */
		stored: ApplicantPortrait | null;
		/** Show that a photo is still needed (after they try to continue). */
		showProblem?: boolean;
		/** The chosen photo is saved in this browser, so it's safe to close the page. */
		kept?: boolean;
		disabled?: boolean;
	};

	let { draft = $bindable(), stored = $bindable(), showProblem = false, kept = false, disabled = false }: Props = $props();

	let error = $state('');
	let inputEl = $state<HTMLInputElement>();

	const missing = $derived(!draft && !stored);
	const previewSrc = $derived(draft?.previewUrl ?? stored?.url ?? '');

	function choose(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';
		if (!file) return;
		const problem = validateImageFile(file);
		if (problem) {
			error = problem;
			return;
		}
		error = '';
		if (draft) URL.revokeObjectURL(draft.previewUrl);
		draft = { file, previewUrl: URL.createObjectURL(file) };
	}

	/** Drops a photo that was only chosen here. A stored one stays on the server until a new one replaces it at submit, so a failed submit never leaves the applicant without one. */
	function remove() {
		error = '';
		if (draft) {
			URL.revokeObjectURL(draft.previewUrl);
			draft = null;
		}
	}
</script>

<div class="space-y-3" data-testid="applicant-portrait">
	<div>
		<p class="t-eyebrow text-accent">Your photo</p>
		<h2 class="t-h3 mt-2">Add a photo of yourself</h2>
		<p class="mt-2 text-sm text-muted-foreground">
			A clear picture of you. It becomes the photo on your public profile, so visitors can see who made the work. You can change it any time later.
		</p>
		{#if kept && draft}
			<p class="mt-2 text-sm text-muted-foreground" data-testid="portrait-kept">Your photo is saved on this device until you submit.</p>
		{/if}
	</div>

	<div class="flex items-center gap-4">
		<div
			class="h-24 w-24 shrink-0 overflow-hidden rounded-full border bg-muted {showProblem && missing ? 'border-destructive' : 'border-border'}"
			aria-hidden={previewSrc ? undefined : 'true'}
		>
			{#if draft}
				<!-- Local preview: not uploaded until you submit. -->
				<img src={draft.previewUrl} alt="Portrait of you" class="h-full w-full object-cover" />
			{:else if stored}
				<BlurhashImage src={stored.url} alt="Portrait of you" blurhash={stored.blurhash} width={96} class="h-full w-full object-cover" />
			{/if}
		</div>
		<div class="space-y-2">
			<input
				bind:this={inputEl}
				type="file"
				accept="image/jpeg,image/png,image/webp"
				class="sr-only"
				id="portrait-input"
				aria-describedby={error ? 'portrait-error' : undefined}
				{disabled}
				onchange={choose}
			/>
			<button
				type="button"
				class="inline-flex h-10 items-center rounded-sm border border-border px-4 text-sm transition hover:bg-muted disabled:opacity-50"
				{disabled}
				onclick={() => inputEl?.click()}
			>
				{draft || stored ? 'Choose a different photo' : 'Choose a photo'}
			</button>
			{#if draft}
				<button type="button" class="block text-sm text-muted-foreground underline underline-offset-4 hover:text-foreground" onclick={remove}>
					Remove
				</button>
			{/if}
		</div>
	</div>

	{#if showProblem && missing && !error}
		<p class="text-sm text-destructive" role="alert">Please add a photo of yourself to continue.</p>
	{/if}
	{#if error}<Alert tone="error"><span id="portrait-error">{error}</span></Alert>{/if}
</div>
