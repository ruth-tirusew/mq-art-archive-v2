<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { authService, currentUser } from '$lib/application/auth';
  import { ErrorBoundary, NavProgress, Toaster } from '$ui';
  import AdminSidebar from '$lib/components/AdminSidebar.svelte';
  import { SidebarProvider, SidebarTrigger } from '$lib/components/ui/sidebar';
  import LayoutDashboard from 'lucide-svelte/icons/layout-dashboard';
  import ClipboardCheck from 'lucide-svelte/icons/clipboard-check';
  import Users from 'lucide-svelte/icons/users';
  import Image from 'lucide-svelte/icons/image';
  import BookOpen from 'lucide-svelte/icons/book-open';
  import Calendar from 'lucide-svelte/icons/calendar';
  import Settings from 'lucide-svelte/icons/settings';
  import UserCog from 'lucide-svelte/icons/user-cog';

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
      groups: []
    },
    {
      id: 'review',
      href: '/applications',
      label: 'Applications',
      icon: ClipboardCheck,
      title: 'Applications',
      match: (pathname) => pathname === '/applications' || pathname.startsWith('/applications/'),
      groups: []
    },
    {
      id: 'artists',
      href: '/artists',
      label: 'Artists',
      icon: Users,
      title: 'Artists',
      match: (pathname) => pathname === '/artists' || pathname.startsWith('/artists/'),
      groups: [
        {
          heading: 'Directory',
          items: [
            { href: '/artists', label: 'All' },
            { href: '/artists?status=pending', label: 'Pending' },
            { href: '/artists?status=approved', label: 'Approved' },
            { href: '/artists?status=draft', label: 'Draft' },
            { href: '/artists/new', label: 'New artist' }
          ]
        }
      ]
    },
    {
      id: 'users',
      href: '/users',
      label: 'Users',
      icon: UserCog,
      title: 'Users',
      match: (pathname) => pathname === '/users',
      groups: []
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
            { href: '/events', label: 'All' },
            { href: '/events?status=pending', label: 'Pending' },
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
    return !urlStatus;
  }

  const activeSection = $derived(
    sections.find((s) => s.match($page.url.pathname, $page.url.searchParams)) ?? sections[0]
  );

  const breadcrumbLeaf = $derived.by(() => {
    const pathname = $page.url.pathname;
    if (pathname.endsWith('/new')) return 'New';
    if (/\/[^/]+$/.test(pathname) && pathname.split('/').length > 2) {
      // Name the record on screen (set by each detail page's loader) rather than a literal "Detail".
      const d = $page.data as Record<string, { display_name?: string; title?: string } | undefined>;
      return (
        d.artist?.display_name ?? d.application?.display_name ?? d.post?.title ?? d.event?.title ?? d.article?.title ?? 'Detail'
      );
    }
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

<NavProgress />
<Toaster />

{#if isAuthRoute}
  <div class="min-h-screen bg-background">
    {@render children()}
  </div>
{:else}
  <SidebarProvider>
    <AdminSidebar
      {topSections}
      {bottomSections}
      activeSectionId={activeSection.id}
      {isItemActive}
      email={$currentUser?.email}
      onlogout={logout}
    />

    <!-- Header + main -->
    <div class="flex min-w-0 flex-1 flex-col">
      <header
        class="flex h-12 shrink-0 items-center gap-2 border-b border-border/50 px-3 md:px-4"
      >
        <SidebarTrigger />
        <span class="mx-1 h-4 w-px bg-border" aria-hidden="true"></span>
        <nav class="flex min-w-0 items-center gap-1.5 text-sm" aria-label="Breadcrumb">
          {#if breadcrumbLeaf !== activeSection.title}
            <a href={activeSection.href} class="truncate text-muted-foreground hover:text-foreground"
              >{activeSection.title}</a
            >
            <span class="text-muted-foreground/50" aria-hidden="true">/</span>
            <span class="truncate text-foreground">{breadcrumbLeaf}</span>
          {:else}
            <span class="truncate text-foreground">{activeSection.title}</span>
          {/if}
        </nav>
      </header>

      <!-- Main content -->
      <main class="page-enter min-h-0 flex-1 overflow-y-auto px-6 py-8 md:px-8 lg:px-10">
        <ErrorBoundary>
          <!-- Remount per path so edit forms never keep the previous record's values. -->
          {#key $page.url.pathname}
            {@render children()}
          {/key}
        </ErrorBoundary>
      </main>
    </div>
  </SidebarProvider>
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
