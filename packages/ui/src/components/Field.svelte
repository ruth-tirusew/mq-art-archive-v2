<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		label: string;
		/** Falls back to a generated id. Pass it to the control via the snippet argument. */
		id?: string;
		hint?: string;
		/** A message here marks the control invalid and is announced to screen readers. */
		error?: string | null;
		required?: boolean;
		children: Snippet<[{ id: string; describedby: string | undefined; invalid: boolean; required: boolean }]>;
	};

	let { label, id, hint, error = null, required = false, children }: Props = $props();

	const uid = $props.id();
	const controlId = $derived(id ?? `f-${uid}`);
	const hintId = $derived(`${controlId}-hint`);
	const errorId = $derived(`${controlId}-error`);
	const describedby = $derived(error ? errorId : hint ? hintId : undefined);
</script>

<div class="flex flex-col gap-1.5">
	<label for={controlId} class="t-eyebrow text-foreground">
		{label}{#if required}<span class="text-accent" aria-hidden="true"> *</span>{/if}
	</label>
	{@render children({ id: controlId, describedby, invalid: !!error, required })}
	{#if error}
		<p id={errorId} role="alert" class="text-[13px] text-destructive">{error}</p>
	{:else if hint}
		<p id={hintId} class="text-[13px] text-muted-foreground">{hint}</p>
	{/if}
</div>
