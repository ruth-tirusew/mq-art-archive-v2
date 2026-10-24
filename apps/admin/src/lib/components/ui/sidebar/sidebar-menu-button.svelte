<script lang="ts">
  import { Tooltip } from 'bits-ui';
  import type { Component } from 'svelte';

  /** Any lucide-svelte icon. */
  type IconComponent = typeof import('lucide-svelte/icons/users').default;
  import { cn } from '$lib/utils';
  import { useSidebar } from './context.svelte';

  let {
    label,
    icon,
    href,
    active = false,
    onclick,
    class: className
  }: {
    label: string;
    icon: IconComponent;
    href?: string;
    active?: boolean;
    onclick?: () => void;
    class?: string;
  } = $props();

  const sidebar = useSidebar();
  const Icon = $derived(icon as IconComponent);

  const classes = $derived(
    cn(
      'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
      sidebar.collapsed && 'h-9 w-9 justify-center p-0',
      active
        ? 'bg-muted font-medium text-foreground'
        : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground',
      className
    )
  );
</script>

{#snippet content()}
  <Icon class="h-4 w-4 shrink-0" strokeWidth={1.5} />
  <span class={sidebar.collapsed ? 'sr-only' : 'truncate'}>{label}</span>
{/snippet}

{#snippet control(props: Record<string, unknown>)}
  {#if href}
    <a {href} aria-current={active ? 'page' : undefined} {...props} class={classes}>
      {@render content()}
    </a>
  {:else}
    <button type="button" {onclick} {...props} class={classes}>
      {@render content()}
    </button>
  {/if}
{/snippet}

{#if sidebar.collapsed}
  <Tooltip.Root>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        {@render control(props)}
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Portal>
      <Tooltip.Content
        side="right"
        sideOffset={10}
        class="z-50 rounded-md bg-foreground px-2.5 py-1.5 text-xs text-background shadow-md"
      >
        {label}
      </Tooltip.Content>
    </Tooltip.Portal>
  </Tooltip.Root>
{:else}
  {@render control({})}
{/if}
