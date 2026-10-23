<script lang="ts">
  import { useSidebar } from './context.svelte';

  const sidebar = useSidebar();
</script>

<!-- A thin hit area on the sidebar's edge; click to dock or expand. -->
<button
  type="button"
  tabindex="-1"
  aria-label="Toggle sidebar"
  title="Toggle sidebar"
  onclick={sidebar.toggle}
  class="absolute inset-y-0 -right-2 z-10 hidden w-4 cursor-w-resize after:absolute after:inset-y-0 after:left-1/2 after:w-px after:bg-transparent after:transition-colors hover:after:bg-border sm:flex {sidebar.collapsed
    ? 'cursor-e-resize'
    : ''}"
></button>
