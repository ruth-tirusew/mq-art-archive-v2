<script lang="ts">
  import { cn } from '$lib/utils';

  let { class: className }: { class?: string } = $props();
</script>

<div class={cn('shrink-0 bg-border h-[1px] w-full', className)} role="separator"></div>
