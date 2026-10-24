<script lang="ts">
  import PanelLeft from 'lucide-svelte/icons/panel-left';
  import { cn } from '$lib/utils';
  import { useSidebar } from './context.svelte';

  let { class: className }: { class?: string } = $props();

  const sidebar = useSidebar();
</script>

<button
  type="button"
  onclick={sidebar.toggle}
  aria-label="Toggle sidebar"
  title="Toggle sidebar (Ctrl/⌘ B)"
  class={cn(
    'inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
    className
  )}
>
  <PanelLeft class="h-4 w-4" strokeWidth={1.5} />
</button>
