<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className, children }: { class?: string; children: import('svelte').Snippet } = $props();
</script>

<div class={cn('relative w-full overflow-auto', className)}>
  <table class="w-full caption-bottom text-sm">
    {@render children()}
  </table>
</div>
