<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, for: htmlFor, children }: {
    class?: string;
    for?: string;
    children: import('svelte').Snippet;
  } = $props();
</script>

<label
  for={htmlFor}
  class={cn('font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground', className)}
>
  {@render children()}
</label>
