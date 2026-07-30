<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, children }: { class?: string; children: import('svelte').Snippet } = $props();
</script>

<td class={cn('p-4 align-middle', className)}>
  {@render children()}
</td>
