<script lang="ts">
  import { Button } from '$lib/components/ui/button';
  import BotanicalMark from '$lib/components/BotanicalMark.svelte';

  let {
    title,
    description,
    actionLabel,
    actionHref
  }: { title: string; description?: string; actionLabel?: string; actionHref?: string } = $props();
</script>

<div
  class="flex flex-col items-center rounded-lg border border-dashed border-border bg-card/40 px-6 py-16 text-center"
>
  <BotanicalMark class="h-10 w-9 text-foreground/25" />
  <p class="mt-4 font-display text-xl text-foreground">{title}</p>
  {#if description}
    <p class="mt-2 max-w-sm text-sm text-muted-foreground">{description}</p>
  {/if}
  {#if actionLabel && actionHref}
    <Button href={actionHref} variant="outline" class="mt-6">{actionLabel}</Button>
  {/if}
</div>
