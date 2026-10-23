<script lang="ts">
  import { Button } from '$lib/components/ui/button';

  let {
    open = $bindable(false),
    title,
    description,
    confirmLabel = 'Confirm',
    destructive = false,
    busy = false,
    onconfirm
  }: {
    open: boolean;
    title: string;
    description?: string;
    confirmLabel?: string;
    destructive?: boolean;
    busy?: boolean;
    onconfirm: () => void | Promise<void>;
  } = $props();

  let dialog = $state<HTMLDialogElement>();

  $effect(() => {
    if (!dialog) return;
    if (open && !dialog.open) dialog.showModal();
    if (!open && dialog.open) dialog.close();
  });
</script>

<dialog
  bind:this={dialog}
  onclose={() => (open = false)}
  class="m-auto w-[min(92vw,26rem)] rounded-lg border border-border bg-card p-6 text-foreground shadow-xl backdrop:bg-foreground/30"
>
  <h2 class="font-display text-xl">{title}</h2>
  {#if description}
    <p class="mt-2 text-sm text-muted-foreground">{description}</p>
  {/if}
  <div class="mt-6 flex justify-end gap-2">
    <Button variant="ghost" disabled={busy} onclick={() => (open = false)}>Cancel</Button>
    <Button variant={destructive ? 'destructive' : 'default'} disabled={busy} onclick={onconfirm}>
      {busy ? 'Working…' : confirmLabel}
    </Button>
  </div>
</dialog>
