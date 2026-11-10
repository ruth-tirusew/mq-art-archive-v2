<script lang="ts">
  import { untrack } from 'svelte';
  import { currentUser } from '$lib/application/auth';
  import { usersService } from '$lib/composition';
  import { ROLES, USERS_PAGE_SIZE, type Role, type User } from '$lib/core/domain/user';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import StaffInvitesPanel from '$lib/components/team/StaffInvitesPanel.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import { Button, ConfirmDialog, toast } from '$ui';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow
  } from '$lib/components/ui/table';
  import { cn } from '$lib/utils';
  import { plural } from '$lib/utils/plural';

  let { data } = $props();

  const roles = ROLES;
  const roleBadge: Record<Role, 'default' | 'success' | 'muted' | 'warning' | 'outline'> = {
    admin: 'default',
    moderator: 'default',
    artist: 'success',
    contributor: 'warning',
    institution: 'outline',
    public: 'muted'
  };
  const roleHelp: Record<Role, string> = {
    public: 'Can read the site, save articles and comment.',
    artist: 'Gets a studio to manage a public profile and works.',
    institution: 'Institution account.',
    contributor: 'Can submit wiki articles through the studio.',
    moderator: 'Reviews and moderates content in this admin panel. Cannot manage users or change system settings.',
    admin: 'Full access to this admin panel, including users and settings.'
  };

  // The first page comes from the route loader; only "load more" and retries fetch here.
  let users = $state<User[]>(untrack(() => data.users));
  let loading = $state(false);
  let error = $state('');
  let loadFailed = $state(false);
  let loadingMore = $state(false);
  let hasMore = $state(untrack(() => data.users.length === USERS_PAGE_SIZE));
  let query = $state('');
  let roleFilter = $state<Role | 'all'>('all');

  let pending = $state<{ user: User; role: Role } | null>(null);
  let confirmOpen = $state(false);
  let saving = $state(false);

  async function load(more = false) {
    error = '';
    loadFailed = false;
    if (more) loadingMore = true;
    else loading = true;
    try {
      const page = await usersService.listPage(more ? users.length : 0);
      users = more ? [...users, ...page] : page;
      hasMore = page.length === USERS_PAGE_SIZE;
    } catch (cause) {
      loadFailed = true;
      error = cause instanceof Error ? cause.message : 'Could not load users';
    } finally {
      loading = false;
      loadingMore = false;
    }
  }

  const counts = $derived.by(() => {
    const result: Record<string, number> = { all: users.length };
    for (const role of roles) result[role] = users.filter((u) => u.role === role).length;
    return result;
  });

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    return users.filter(
      (u) =>
        (roleFilter === 'all' || u.role === roleFilter) &&
        (!q || u.email.toLowerCase().includes(q) || (u.display_name ?? '').toLowerCase().includes(q))
    );
  });

  function requestChange(user: User, role: Role) {
    if (role === user.role) return;
    pending = { user, role };
    confirmOpen = true;
  }

  async function applyChange() {
    if (!pending) return;
    saving = true;
    error = '';
    try {
      const updated = await usersService.changeRole(pending.user.id, pending.role);
      users = users.map((item) => (item.id === updated.id ? updated : item));
      toast.success(`${updated.email} is now ${updated.role}.`);
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not change role';
    } finally {
      saving = false;
      confirmOpen = false;
      pending = null;
    }
  }
</script>

<svelte:head><title>Users — Artiv admin</title></svelte:head>

<PageHeader
  eyebrow="Users"
  title="Users"
  description="Search accounts and manage who has access to what. Role changes take effect immediately."
/>

<StaffInvitesPanel />

{#if error}
  <Alert variant="destructive" class="mb-4 flex items-center justify-between gap-4">
    <span>{error}</span>
    {#if loadFailed}
      <Button variant="secondary" size="sm" onclick={() => load(users.length > 0)}>Try again</Button>
    {/if}
  </Alert>
{/if}

<div class="mb-4 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
  <div class="flex gap-2 overflow-x-auto" role="group" aria-label="Filter by role">
    {#each ['all', ...roles] as option}
      <button
        type="button"
        aria-pressed={roleFilter === option}
        onclick={() => (roleFilter = option as Role | 'all')}
        class={cn(
          'shrink-0 rounded-md border px-3 py-1 text-xs capitalize transition-colors',
          roleFilter === option
            ? 'border-foreground bg-muted text-foreground'
            : 'border-border text-muted-foreground hover:text-foreground'
        )}
      >
        {option}
        <span class="ml-1 text-muted-foreground">{counts[option] ?? 0}</span>
      </button>
    {/each}
  </div>
  <input
    type="search"
    bind:value={query}
    placeholder="Search by name or email"
    aria-label="Search users"
    class="h-9 w-full rounded-md border border-input bg-card px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring md:w-72"
  />
</div>

{#if loading}
  <div class="space-y-2" role="status" aria-live="polite">
    <span class="sr-only">Loading users</span>
    <Skeleton class="h-14 w-full" />
    <Skeleton class="h-14 w-full" />
    <Skeleton class="h-14 w-full" />
    <Skeleton class="h-14 w-full" />
  </div>
{:else if filtered.length === 0 && users.length === 0 && loadFailed}
  <!-- load failed: the error above has the retry; don't also claim there are no users -->
{:else if filtered.length === 0}
  <EmptyState
    title="No users found"
    description="Try a different search or role filter."
  />
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>User</TableHead>
        <TableHead>Role</TableHead>
        <TableHead class="w-44 text-right">Change role</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each filtered as user}
        {@const isSelf = user.id === $currentUser?.id}
        <TableRow>
          <TableCell>
            <p class="font-medium text-foreground">
              {user.display_name || user.email}
              {#if isSelf}<span class="ml-1 text-xs font-normal text-muted-foreground">(you)</span>{/if}
            </p>
            {#if user.display_name}
              <p class="text-xs text-muted-foreground">{user.email}</p>
            {/if}
          </TableCell>
          <TableCell><Badge variant={roleBadge[user.role]}>{user.role}</Badge></TableCell>
          <TableCell class="text-right">
            <select
              class="h-8 rounded-md border border-input bg-card px-2 text-sm disabled:opacity-50"
              value={user.role}
              disabled={isSelf}
              title={isSelf ? 'You cannot change your own role' : undefined}
              aria-label="Change role for {user.email}"
              onchange={(event) => {
                requestChange(user, event.currentTarget.value as Role);
                event.currentTarget.value = user.role;
              }}
            >
              {#each roles as role}<option value={role}>{role}</option>{/each}
            </select>
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
  <p class="mt-3 text-xs text-muted-foreground">
    Showing {plural(filtered.length, 'user')}{hasMore ? ` of the first ${users.length} loaded` : ''}.
  </p>
  {#if hasMore}
    <div class="mt-3">
      <Button variant="secondary" size="sm" loading={loadingMore} onclick={() => load(true)}>Load more users</Button>
    </div>
  {/if}
{/if}

<ConfirmDialog
  bind:open={confirmOpen}
  title={pending ? `Make ${pending.user.display_name || pending.user.email} ${pending.role}?` : 'Change role'}
  description={pending
    ? `${roleHelp[pending.role]}${pending.user.role === 'admin' ? ' They will lose admin access.' : ''}`
    : undefined}
  confirmLabel="Change role"
  destructive={pending?.role === 'admin' || pending?.user.role === 'admin'}
  busy={saving}
  onconfirm={applyChange}
/>
