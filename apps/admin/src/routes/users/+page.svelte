<script lang="ts">
  import { onMount } from 'svelte';
  import { apiFetch } from '$lib/adapters/api/client';
  import { currentUser } from '$lib/application/auth';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import { ConfirmDialog } from '$ui';
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

  type Role = 'public' | 'artist' | 'institution' | 'contributor' | 'admin';
  type User = { id: string; email: string; role: Role; display_name?: string };

  const roles: Role[] = ['public', 'artist', 'institution', 'contributor', 'admin'];
  const roleBadge: Record<Role, 'default' | 'success' | 'muted' | 'warning' | 'outline'> = {
    admin: 'default',
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
    admin: 'Full access to this admin panel.'
  };

  let users = $state<User[]>([]);
  let loading = $state(true);
  let error = $state('');
  let notice = $state('');
  let query = $state('');
  let roleFilter = $state<Role | 'all'>('all');

  let pending = $state<{ user: User; role: Role } | null>(null);
  let confirmOpen = $state(false);
  let saving = $state(false);

  onMount(async () => {
    try {
      users = await apiFetch<User[]>('/admin/v1/users?limit=200');
    } catch (cause) {
      error = cause instanceof Error ? cause.message : 'Could not load users';
    } finally {
      loading = false;
    }
  });

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
    notice = '';
    try {
      const updated = await apiFetch<User>(`/admin/v1/users/${pending.user.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ role: pending.role })
      });
      users = users.map((item) => (item.id === updated.id ? updated : item));
      notice = `${updated.email} is now ${updated.role}.`;
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

{#if notice}<Alert class="mb-4">{notice}</Alert>{/if}
{#if error}<Alert variant="destructive" class="mb-4">{error}</Alert>{/if}

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
    Showing {plural(filtered.length, 'user')}{users.length >= 200 ? ' (first 200 loaded)' : ''}.
  </p>
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
