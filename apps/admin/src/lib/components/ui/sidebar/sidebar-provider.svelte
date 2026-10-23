<script lang="ts">
  import { onMount } from 'svelte';
  import { Tooltip } from 'bits-ui';
  import { setSidebar } from './context.svelte';

  let { children }: { children: import('svelte').Snippet } = $props();

  const sidebar = setSidebar();

  onMount(() => sidebar.init());

  // Cmd/Ctrl + B toggles the sidebar, like shadcn's sidebar block.
  function onkeydown(event: KeyboardEvent) {
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'b') {
      event.preventDefault();
      sidebar.toggle();
    }
  }
</script>

<svelte:window {onkeydown} />

<Tooltip.Provider delayDuration={0} skipDelayDuration={0}>
  <div class="flex h-screen overflow-hidden bg-background" data-sidebar-state={sidebar.state}>
    {@render children()}
  </div>
</Tooltip.Provider>
