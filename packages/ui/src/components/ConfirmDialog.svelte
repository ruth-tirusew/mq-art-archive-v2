<script lang="ts">
	import Button from './Button.svelte';

	type Props = {
		open: boolean;
		title: string;
		description?: string;
		confirmLabel?: string;
		cancelLabel?: string;
		/** Use for irreversible or public-facing actions (delete, unpublish, reject). */
		destructive?: boolean;
		/** Disables both buttons and shows a spinner while the action runs. */
		busy?: boolean;
		onconfirm: () => void | Promise<void>;
		/** Called after the dialog closes without confirming. Closing itself is handled here. */
		oncancel?: () => void;
	};

	let {
		open = $bindable(false),
		title,
		description,
		confirmLabel = 'Confirm',
		cancelLabel = 'Cancel',
		destructive = false,
		busy = false,
		onconfirm,
		oncancel
	}: Props = $props();

	let dialog: HTMLDialogElement | undefined = $state();

	function cancel() {
		if (busy) return;
		open = false;
		oncancel?.();
	}

	$effect(() => {
		if (!dialog) return;
		if (open && !dialog.open) dialog.showModal();
		if (!open && dialog.open) dialog.close();
	});
</script>

<dialog
	bind:this={dialog}
	oncancel={(e) => {
		e.preventDefault();
		cancel();
	}}
	onclose={() => (open = false)}
	onclick={(e) => {
		if (e.target === dialog) cancel();
	}}
	aria-labelledby="confirm-title"
	class="m-auto w-[min(28rem,calc(100vw-2rem))] rounded-sm border border-border bg-background p-0 text-foreground shadow-xl backdrop:bg-ink/50 backdrop:backdrop-blur-[2px]"
>
	<div class="flex flex-col gap-4 p-6 md:p-8">
		<p class="t-eyebrow {destructive ? 'text-destructive' : 'text-accent'}">
			{destructive ? 'Are you sure' : 'Confirm'}
		</p>
		<h2 id="confirm-title" class="t-h3">{title}</h2>
		{#if description}<p class="t-body text-muted-foreground">{description}</p>{/if}
		<div class="mt-2 flex justify-end gap-3">
			<Button variant="ghost" disabled={busy} onclick={cancel}>{cancelLabel}</Button>
			<Button variant={destructive ? 'destructive' : 'primary'} loading={busy} onclick={onconfirm}>
				{confirmLabel}
			</Button>
		</div>
	</div>
</dialog>
