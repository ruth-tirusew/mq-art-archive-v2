<script lang="ts">
  import { cn } from '$lib/utils';
  import type { HTMLAttributes } from 'svelte/elements';

  let {
    class: className,
    children,
    ...rest
  }: HTMLAttributes<HTMLTableRowElement> & {
    children: import('svelte').Snippet;
  } = $props();
</script>

<tr class={cn('border-b border-border/70 transition-colors hover:bg-muted/50', className)} {...rest}>
  {@render children()}
</tr>
