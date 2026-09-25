<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { authService, currentUser } from '$lib/application/auth';
  import { Button } from '$lib/components/ui/button';
  import { cn } from '$lib/utils';
  import LayoutDashboard from 'lucide-svelte/icons/layout-dashboard';
  import ClipboardCheck from 'lucide-svelte/icons/clipboard-check';
  import Users from 'lucide-svelte/icons/users';
  import Image from 'lucide-svelte/icons/image';
  import BookOpen from 'lucide-svelte/icons/book-open';
  import Calendar from 'lucide-svelte/icons/calendar';
  import Settings from 'lucide-svelte/icons/settings';

  let { children } = $props();

  type NavItem = { href: string; label: string };
  type NavGroup = { heading?: string; items: NavItem[] };
  type PrimarySection = {
    id: string;
    href: string;
    label: string;
    icon: typeof LayoutDashboard;
    title: string;
    groups: NavGroup[];
    position?: 'top' | 'bottom';
    match: (pathname: string, search: URLSearchParams) => boolean;
  };

  const sections: PrimarySection[] = [
    {
      id: 'overview',
      href: '/',
      label: 'Overview',
      icon: LayoutDashboard,
      title: 'Overview',
      match: (pathname) => pathname === '/',
      groups: [{ items: [{ href: '/', label: 'Dashboard' }] }]
    },
    {
      id: 'review',
      href: '/applications',
      label: 'Review',
      icon: ClipboardCheck,
      title: 'Review',
      match: (pathname) => pathname === '/applications' || pathname.startsWith('/applications/'),
      groups: [{ items: [{ href: '/applications', label: 'Applications' }] }]
    },
    {
      id: 'artists',
      href: '/artists',
      label: 'Artists',
      icon: Users,
      title: 'Artists',
      match: (pathname) => pathname === '/users' || pathname === '/artists' || pathname.startsWith('/artists/'),
      groups: [
        {
          heading: 'Directory',
          items: [
            { href: '/artists', label: 'All' },
            { href: '/artists?status=pending', label: 'Pending' },
            { href: '/artists?status=approved', label: 'Approved' },
            { href: '/artists?status=draft', label: 'Draft' },
            { href: '/users', label: 'Users' },
            { href: '/artists/new', label: 'New artist' }
          ]
        }
      ]
    },
    {
      id: 'content',
      href: '/posts',
      label: 'Content',
      icon: Image,
      title: 'Content',
      match: (pathname) => pathname === '/posts' || pathname.startsWith('/posts/'),
      groups: [
        {
          heading: 'Posts',
          items: [
            { href: '/posts', label: 'All' },
            { href: '/posts?status=draft', label: 'Drafts' },
            { href: '/posts?status=published', label: 'Published' },
            { href: '/posts?status=archived', label: 'Archived' },
            { href: '/posts/new', label: 'New post' }
          ]
        }
      ]
    },
    {
      id: 'wiki',
      href: '/wiki',
      label: 'Wiki',
      icon: BookOpen,
      title: 'Wiki',
      match: (pathname) => pathname === '/wiki' || pathname.startsWith('/wiki/'),
      groups: [
        {
          heading: 'Articles',
          items: [
            { href: '/wiki', label: 'All' },
            { href: '/wiki?status=draft', label: 'Drafts' },
            { href: '/wiki?status=published', label: 'Published' },
            { href: '/wiki?status=archived', label: 'Archived' },
            { href: '/wiki/submissions', label: 'Submissions' },
            { href: '/wiki/new', label: 'New article' }
          ]
        }
      ]
    },
    {
      id: 'events',
      href: '/events',
      label: 'Events',
      icon: Calendar,
      title: 'Events',
      match: (pathname) => pathname === '/events' || pathname.startsWith('/events/'),
      groups: [
        {
          heading: 'Calendar',
          items: [
            { href: '/events?status=all', label: 'All' },
            { href: '/events', label: 'Pending' },
            { href: '/events?status=approved', label: 'Approved' },
            { href: '/events?status=rejected', label: 'Rejected' },
            { href: '/events/new', label: 'New event' }
          ]
        }
      ]
    },
    {
      id: 'settings',
      href: '/settings',
      label: 'Settings',
      icon: Settings,
      title: 'Settings',
      position: 'bottom',
      match: (pathname) => pathname === '/settings' || pathname.startsWith('/settings/'),
      groups: [
        {
          heading: 'Configuration',
          items: [
            { href: '/settings#profile', label: 'Profile' },
            { href: '/settings#account', label: 'Account' },
            { href: '/settings#notifications', label: 'Notifications' }
          ]
        },
        {
          heading: 'Integrations',
          items: [{ href: '/settings#integrations', label: 'Scrape & Telegram' }]
        }
      ]
    }
  ];

  const topSections = sections.filter((s) => s.position !== 'bottom');
  const bottomSections = sections.filter((s) => s.position === 'bottom');

  onMount(async () => {
    const path = $page.url.pathname;
    if (path === '/login' || path.startsWith('/auth/')) {
      await authService.load();
      return;
    }

    const user = await authService.load();
    if (!user) {
      await goto('/login');
      return;
    }
    if (user.role !== 'admin') {
      await goto('/login?error=access_denied');
    }
  });

  async function logout() {
    await authService.logout();
    await goto('/login');
  }

  function isItemActive(href: string) {
    const [pathAndQuery, hashPart] = href.split('#');
    const [pathname, queryPart] = pathAndQuery.split('?');
    const url = $page.url;
    const hash = hashPart ?? '';
    const hrefStatus = queryPart ? new URLSearchParams(queryPart).get('status') : null;
    const urlStatus = url.searchParams.get('status');

    if (hash && url.pathname === pathname) {
      const currentHash = url.hash.replace(/^#/, '') || 'profile';
      return currentHash === hash;
    }

    if (pathname.endsWith('/new')) {
      return url.pathname === pathname;
    }

    if (pathname === '/') return url.pathname === '/';

    const onListOrDetail =
      url.pathname === pathname ||
      (url.pathname.startsWith(pathname + '/') && !url.pathname.startsWith(pathname + '/new'));

    if (!onListOrDetail) return false;

    // Exact status filter match (e.g. ?status=draft)
    if (hrefStatus) {
      return url.pathname === pathname && urlStatus === hrefStatus;
    }

    // Default list link (no status param)
    if (pathname === '/events') {
      // Events defaults to pending when status is omitted
      return url.pathname === pathname && (!urlStatus || urlStatus === 'pending');
    }

    return url.pathname === pathname ? !urlStatus : !urlStatus;
  }

  const activeSection = $derived(
    sections.find((s) => s.match($page.url.pathname, $page.url.searchParams)) ?? sections[0]
  );

  // Whether the active section has more than one sub-item worth showing as its own
  // mobile filter row (a section with a single "view everything" item doesn't).
  const activeSectionHasSubItems = $derived(
    activeSection.groups.reduce((count, group) => count + group.items.length, 0) > 1
  );

  const breadcrumbLeaf = $derived.by(() => {
    const pathname = $page.url.pathname;
    if (pathname.endsWith('/new')) return 'New';
    if (/\/[^/]+$/.test(pathname) && pathname.split('/').length > 2) return 'Detail';
    const hash = $page.url.hash.replace(/^#/, '');
    if (pathname === '/settings' && hash) {
      return hash.charAt(0).toUpperCase() + hash.slice(1);
    }
    return activeSection.title;
  });

  const isAuthRoute = $derived(
    $page.url.pathname === '/login' || $page.url.pathname.startsWith('/auth/')
  );
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link
    href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,400;9..144,600&family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap"
    rel="stylesheet"
  />
</svelte:head>

{#if isAuthRoute}
  <div class="min-h-screen bg-background">
    {@render children()}
  </div>
{:else}
  <div class="flex h-screen overflow-hidden bg-background">
    <!-- Sidebar: top-level sections, with the active section's sub-items expanded inline -->
    <aside
      class="hidden w-56 shrink-0 flex-col overflow-y-auto border-r border-border/60 bg-sidebar lg:flex"
      aria-label="Primary"
    >
      <a
        href="/"
        class="px-5 pb-2 pt-5 font-display text-lg tracking-tight text-foreground"
        aria-label="artiv. home"
      >
        artiv.
      </a>

      {#snippet sectionNav(section: PrimarySection)}
        <a
          href={section.href}
          aria-current={activeSection.id === section.id ? 'page' : undefined}
          class={cn(
            'flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
            activeSection.id === section.id
              ? 'bg-muted font-medium text-foreground'
              : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
          )}
        >
          <section.icon class="h-4 w-4 shrink-0" strokeWidth={1.5} />
          {section.label}
        </a>
        {#if activeSection.id === section.id}
          <div class="mb-1 ml-4 flex flex-col gap-3 border-l border-border/60 py-1 pl-3">
            {#each section.groups as group}
              <div class="flex flex-col gap-0.5">
                {#if group.heading}
                  <p
                    class="px-2 pb-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                  >
                    {group.heading}
                  </p>
                {/if}
                {#each group.items as item}
                  <a
                    href={item.href}
                    class={cn(
                      'rounded-md px-2 py-1 text-sm transition-colors',
                      isItemActive(item.href)
                        ? 'bg-muted font-medium text-foreground'
                        : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'
                    )}
                  >
                    {item.label}
                  </a>
                {/each}
              </div>
            {/each}
          </div>
        {/if}
      {/snippet}

      <nav class="flex flex-1 flex-col gap-0.5 px-3 pb-3 pt-2">
        {#each topSections as section}
          {@render sectionNav(section)}
        {/each}
      </nav>

      <nav class="flex flex-col gap-0.5 border-t border-border/60 px-3 py-3">
        {#each bottomSections as section}
          {@render sectionNav(section)}
        {/each}
      </nav>
    </aside>

    <!-- Header + main -->
    <div class="flex min-w-0 flex-1 flex-col">
      <header
        class="flex h-12 shrink-0 items-center justify-between gap-4 border-b border-border/50 px-4 md:px-5"
      >
        <nav class="flex min-w-0 items-center gap-1.5 text-sm" aria-label="Breadcrumb">
          <a href="/" class="shrink-0 font-display text-base tracking-tight text-foreground lg:hidden"
            >artiv.</a
          >
          <span class="text-muted-foreground/50 lg:hidden" aria-hidden="true">/</span>
          <span class="truncate text-muted-foreground">{activeSection.title}</span>
          {#if breadcrumbLeaf !== activeSection.title}
            <span class="text-muted-foreground/50" aria-hidden="true">/</span>
            <span class="truncate text-foreground">{breadcrumbLeaf}</span>
          {/if}
        </nav>

        <div class="flex shrink-0 items-center gap-3">
          {#if $currentUser}
            <span class="hidden text-sm text-muted-foreground sm:inline">{$currentUser.email}</span>
            <Button variant="outline" size="sm" onclick={logout}>Log out</Button>
          {:else}
            <Button variant="outline" size="sm" href="/login">Login</Button>
          {/if}
        </div>
      </header>

      <!-- Mobile nav: top-level sections, then the active section's sub-items -->
      <div class="flex flex-col gap-2 border-b border-border/50 px-3 py-2 lg:hidden">
        <div class="flex gap-2 overflow-x-auto">
          {#each sections as section}
            <a
              href={section.href}
              class={cn(
                'flex shrink-0 items-center gap-1.5 rounded-md border px-3 py-1 text-xs',
                activeSection.id === section.id
                  ? 'border-foreground bg-muted text-foreground'
                  : 'border-border text-muted-foreground'
              )}
            >
              <section.icon class="h-3.5 w-3.5" strokeWidth={1.5} />
              {section.label}
            </a>
          {/each}
        </div>
        {#if activeSectionHasSubItems}
          <div class="flex gap-2 overflow-x-auto">
            {#each activeSection.groups as group}
              {#each group.items as item}
                <a
                  href={item.href}
                  class={cn(
                    'shrink-0 rounded-md border px-3 py-1 text-xs',
                    isItemActive(item.href)
                      ? 'border-foreground bg-muted text-foreground'
                      : 'border-border text-muted-foreground'
                  )}
                >
                  {item.label}
                </a>
              {/each}
            {/each}
          </div>
        {/if}
      </div>

      <!-- Main content -->
      <main class="page-enter min-h-0 flex-1 overflow-y-auto px-6 py-8 md:px-8 lg:px-10">
        {@render children()}
      </main>
    </div>
  </div>
{/if}

<style>
  :global(.page-enter) {
    animation: page-fade-in 0.35s ease-out;
  }

  @keyframes page-fade-in {
    from {
      opacity: 0;
      transform: translateY(6px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
