<script lang="ts" module>
  import type { Component } from 'svelte';

  /** Any lucide-svelte icon. */
  export type IconComponent = typeof import('lucide-svelte/icons/users').default;

  export type NavItem = { href: string; label: string };
  export type NavGroup = { heading?: string; items: NavItem[] };
  export type NavSection = {
    id: string;
    href: string;
    label: string;
    icon: IconComponent;
    groups: NavGroup[];
  };
</script>

<script lang="ts">
  import { Accordion } from 'bits-ui';
  import ChevronRight from 'lucide-svelte/icons/chevron-right';
  import LogOut from 'lucide-svelte/icons/log-out';
  import { Sidebar, SidebarMenuButton, SidebarRail, useSidebar } from '$lib/components/ui/sidebar';
  import { cn } from '$lib/utils';

  let {
    topSections,
    bottomSections,
    activeSectionId,
    isItemActive,
    email,
    onlogout
  }: {
    topSections: NavSection[];
    bottomSections: NavSection[];
    activeSectionId: string;
    isItemActive: (href: string) => boolean;
    email?: string;
    onlogout: () => void;
  } = $props();

  const sidebar = useSidebar();

  // One section open at a time. It follows the route, but can be toggled by hand.
  let openId = $state('');
  $effect(() => {
    openId = activeSectionId;
  });

  const initial = $derived((email ?? '?').charAt(0).toUpperCase());

  function closeMobile() {
    if (sidebar.isMobile) sidebar.openMobile = false;
  }
</script>

{#snippet section(item: NavSection)}
  {#if item.groups.length === 0 || sidebar.collapsed}
    <SidebarMenuButton
      href={item.href}
      label={item.label}
      icon={item.icon}
      active={activeSectionId === item.id}
      onclick={closeMobile}
    />
  {:else}
    {@const Icon = item.icon as IconComponent}
    <Accordion.Item value={item.id}>
      <Accordion.Header>
        <Accordion.Trigger
          class={cn(
            'group flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
            activeSectionId === item.id
              ? 'font-medium text-foreground'
              : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
          )}
        >
          <Icon class="h-4 w-4 shrink-0" strokeWidth={1.5} />
          <span class="truncate">{item.label}</span>
          <ChevronRight
            class="ml-auto h-3.5 w-3.5 shrink-0 text-muted-foreground transition-transform duration-200 group-data-[state=open]:rotate-90"
            strokeWidth={1.5}
          />
        </Accordion.Trigger>
      </Accordion.Header>
      <Accordion.Content
        class="overflow-hidden data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down"
      >
        <div class="ml-4 mt-0.5 flex flex-col gap-0.5 border-l border-border/60 py-1 pl-3">
          {#each item.groups as group}
            <div class="flex flex-col gap-0.5">
              {#each group.items as link}
                <a
                  href={link.href}
                  onclick={closeMobile}
                  aria-current={isItemActive(link.href) ? 'page' : undefined}
                  class={cn(
                    'rounded-md px-2 py-1 text-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
                    isItemActive(link.href)
                      ? 'bg-muted font-medium text-foreground'
                      : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
                  )}
                >
                  {link.label}
                </a>
              {/each}
            </div>
          {/each}
        </div>
      </Accordion.Content>
    </Accordion.Item>
  {/if}
{/snippet}

<Sidebar>
  <!-- Header -->
  <div class={cn('flex h-12 shrink-0 items-center border-b border-border/50', sidebar.collapsed ? 'justify-center' : 'px-5')}>
    <a href="/" class="font-display text-lg tracking-tight text-foreground" aria-label="artiv. home">
      {sidebar.collapsed ? 'a.' : 'artiv.'}
    </a>
  </div>

  <!-- Content -->
  <nav class={cn('flex min-h-0 flex-1 flex-col overflow-y-auto py-3', sidebar.collapsed ? 'items-center px-2' : 'px-3')}>
    <Accordion.Root type="single" bind:value={openId} class="flex w-full flex-col gap-0.5">
      {#each topSections as item}
        {@render section(item)}
      {/each}
    </Accordion.Root>
  </nav>

  <!-- Footer -->
  <div class={cn('shrink-0 border-t border-border/60 py-3', sidebar.collapsed ? 'flex flex-col items-center gap-1 px-2' : 'px-3')}>
    <Accordion.Root type="single" bind:value={openId} class="flex w-full flex-col gap-0.5">
      {#each bottomSections as item}
        {@render section(item)}
      {/each}
    </Accordion.Root>

    {#if sidebar.collapsed}
      <SidebarMenuButton label="Log out" icon={LogOut} onclick={onlogout} />
    {:else}
      <div class="mt-2 flex items-center gap-2 rounded-md px-2 py-1.5">
        <span
          class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-foreground font-mono text-xs text-cream"
          aria-hidden="true"
        >
          {initial}
        </span>
        <div class="min-w-0 flex-1">
          <p class="truncate text-xs text-foreground" title={email}>{email ?? 'Not signed in'}</p>
          <p class="text-[11px] text-muted-foreground">Admin</p>
        </div>
        <button
          type="button"
          onclick={onlogout}
          aria-label="Log out"
          title="Log out"
          class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <LogOut class="h-4 w-4" strokeWidth={1.5} />
        </button>
      </div>
    {/if}
  </div>

  <SidebarRail />
</Sidebar>
