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
  <!--
    The one fixed frame of the admin: exactly the visible viewport tall (dvh, so mobile browser toolbars
    don't push it taller than the screen), clipped, and a positioning context. The only thing that scrolls is
    <main> inside it. `relative` matters: without it, position:absolute descendants (every .sr-only label)
    anchor to the window instead, escape the clipping and give the page a second scrollbar.
  -->
  <div class="relative flex h-dvh overflow-hidden bg-background" data-sidebar-state={sidebar.state}>
    {@render children()}
  </div>
</Tooltip.Provider>
