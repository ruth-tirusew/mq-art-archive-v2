<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, children }: { class?: string; children: import('svelte').Snippet } = $props();
</script>

<h3 class={cn('font-display text-2xl leading-none tracking-tight', className)}>
  {@render children()}
</h3>
