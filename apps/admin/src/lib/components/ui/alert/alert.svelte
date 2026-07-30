<script lang="ts">
  import { cn } from '$lib/utils';

  type Variant = 'default' | 'destructive';

  let {
    variant = 'default',
    class: className,
    children
  }: { variant?: Variant; class?: string; children: import('svelte').Snippet } = $props();

  const variants: Record<Variant, string> = {
    default: 'bg-muted text-foreground border-border',
    destructive: 'border-destructive/50 text-destructive bg-destructive/10'
  };
</script>

<div
  role="alert"
  class={cn('relative w-full rounded-sm border px-4 py-3 text-sm', variants[variant], className)}
>
  {@render children()}
</div>
