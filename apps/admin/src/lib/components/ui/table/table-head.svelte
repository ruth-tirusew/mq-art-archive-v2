<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, children }: { class?: string; children: import('svelte').Snippet } = $props();
</script>

<th
  class={cn(
    'h-10 px-4 text-left align-middle font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground',
    className
  )}
>
  {@render children()}
</th>
