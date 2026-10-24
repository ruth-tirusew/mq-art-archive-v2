<script lang="ts">
  import { Dialog } from 'bits-ui';
  import { cn } from '$lib/utils';
  import { useSidebar } from './context.svelte';

  let {
    class: className,
    children
  }: { class?: string; children: import('svelte').Snippet } = $props();

  const sidebar = useSidebar();
</script>

{#if sidebar.isMobile}
  <Dialog.Root bind:open={sidebar.openMobile}>
    <Dialog.Portal>
      <Dialog.Overlay
        class="fixed inset-0 z-40 bg-foreground/30 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:animate-in data-[state=open]:fade-in-0"
      />
      <Dialog.Content
        class={cn(
          'fixed inset-y-0 left-0 z-50 flex w-72 max-w-[85vw] flex-col border-r border-border/60 bg-sidebar shadow-xl outline-none',
          'data-[state=closed]:animate-out data-[state=closed]:slide-out-to-left data-[state=open]:animate-in data-[state=open]:slide-in-from-left',
          className
        )}
      >
        <Dialog.Title class="sr-only">Navigation</Dialog.Title>
        <Dialog.Description class="sr-only">Main admin navigation</Dialog.Description>
        {@render children()}
      </Dialog.Content>
    </Dialog.Portal>
  </Dialog.Root>
{:else}
  <aside
    aria-label="Primary"
    data-state={sidebar.state}
    data-collapsible={sidebar.collapsed ? 'icon' : ''}
    class={cn(
      'group/sidebar relative flex shrink-0 flex-col border-r border-border/60 bg-sidebar transition-[width] duration-200 ease-linear',
      sidebar.collapsed ? 'w-[3.5rem]' : 'w-60',
      className
    )}
  >
    {@render children()}
  </aside>
{/if}
