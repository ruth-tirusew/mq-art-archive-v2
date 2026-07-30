<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, children }: { class?: string; children: import('svelte').Snippet } = $props();
</script>

<div class={cn('rounded-sm border border-border/70 bg-card text-card-foreground shadow-sm', className)}>
  {@render children()}
</div>
