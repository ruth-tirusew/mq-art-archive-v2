<script lang="ts">
  import { ConfirmDialog, Field, Input, Select, createConfirm, toast } from '$ui';
  import { invalidateAll } from '$app/navigation';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
  import { invitesService } from '$lib/application/invites';
  import type { Invite, InviteCreated, InviteStatus } from '$lib/core/domain/invite';

  let { data } = $props();

  const confirm = createConfirm();

  let email = $state('');
  let displayName = $state('');
  let handle = $state('');
  let ttlDays = $state('14');
  let submitting = $state(false);
  let errorMessage = $state('');
  let fieldErrors = $state<{ email?: string; handle?: string }>({});
  /** The link is only ever returned once, so it is held here until the admin dismisses it. */
  let created = $state<InviteCreated | null>(null);

  const statusVariant: Record<InviteStatus, 'warning' | 'success' | 'muted' | 'destructive'> = {
    live: 'warning',
    claimed: 'success',
    expired: 'muted',
    revoked: 'destructive'
  };
  const statusLabel: Record<InviteStatus, string> = {
    live: 'Waiting',
    claimed: 'Claimed',
    expired: 'Expired',
    revoked: 'Revoked'
  };

  const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  const HANDLE = /^[a-z0-9_]{3,30}$/;

  function formatDate(value?: string) {
    if (!value) return '—';
    return new Date(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function validate(): boolean {
    fieldErrors = {};
    if (!EMAIL.test(email.trim())) fieldErrors.email = 'Enter a valid email address.';
    const h = handle.trim();
    if (h && !HANDLE.test(h)) fieldErrors.handle = 'Use 3–30 lowercase letters, numbers or underscores.';
    return Object.keys(fieldErrors).length === 0;
  }

  async function create(event: Event) {
    event.preventDefault();
    errorMessage = '';
    if (!validate()) return;
    submitting = true;
    try {
      created = await invitesService.create({
        email: email.trim(),
        display_name: displayName.trim() || undefined,
        handle: handle.trim() || undefined,
        ttl_days: Number(ttlDays)
      });
      email = displayName = handle = '';
      await invalidateAll();
      toast.success('Invite created.');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Could not create the invite';
    } finally {
      submitting = false;
    }
  }

  async function copyLink() {
    if (!created) return;
    try {
      await navigator.clipboard.writeText(created.link);
      toast.success('Link copied.');
    } catch {
      toast.error('Could not copy. Select the link and copy it by hand.');
    }
  }

  function askResend(invite: Invite) {
    confirm.ask({
      title: `Send ${invite.email} a new link?`,
      description: 'The old link stops working immediately. Their name and handle carry over.',
      confirmLabel: 'Send new link',
      run: async () => {
        try {
          created = await invitesService.resend(invite.id);
          await invalidateAll();
          toast.success('New link created.');
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : 'Could not resend the invite';
        }
      }
    });
  }

  function askRevoke(invite: Invite) {
    confirm.ask({
      title: `Revoke the invite for ${invite.email}?`,
      description: 'The link stops working and the reserved handle is released.',
      confirmLabel: 'Revoke',
      destructive: true,
      run: async () => {
        try {
          await invitesService.revoke(invite.id);
          await invalidateAll();
          toast.success('Invite revoked.');
        } catch (err) {
          errorMessage = err instanceof Error ? err.message : 'Could not revoke the invite';
        }
      }
    });
  }
</script>

<svelte:head>
  <title>Invites — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Onboarding"
  title="Artist invites"
  description="Invite artists you already know. They answer the questionnaire and apply as usual; you still approve every application."
/>

{#if errorMessage}
  <Alert variant="destructive" class="mb-6">{errorMessage}</Alert>
{/if}

{#if created}
  <div data-testid="invite-link" class="mb-6">
  <Card class="border-accent/50">
    <CardHeader>
      <CardTitle>Invite link for {created.invite.email}</CardTitle>
    </CardHeader>
    <CardContent class="space-y-3">
      <p class="text-sm text-muted-foreground">
        This is the only time the link is shown. We also emailed it, but if mail isn't set up, copy it now and send it yourself.
        They must sign in with <strong>{created.invite.email}</strong> and have verified that address.
      </p>
      <div class="flex flex-col gap-2 sm:flex-row">
        <input
          readonly
          aria-label="Invite link"
          value={created.link}
          onfocus={(e) => e.currentTarget.select()}
          class="h-9 flex-1 rounded-md border border-border bg-background px-3 font-mono text-xs"
        />
        <Button variant="secondary" onclick={copyLink}>Copy link</Button>
        <Button variant="ghost" onclick={() => (created = null)}>Done</Button>
      </div>
    </CardContent>
  </Card>
  </div>
{/if}

<Card class="mb-8">
  <CardHeader>
    <CardTitle>New invite</CardTitle>
  </CardHeader>
  <CardContent>
    <form class="grid gap-4 md:grid-cols-2" onsubmit={create} novalidate>
      <Field label="Email" required error={fieldErrors.email} hint="The address the invite is for. Only this account can claim it.">
        {#snippet children({ id, describedby, invalid })}
          <Input {id} type="email" autocomplete="off" aria-describedby={describedby} {invalid} bind:value={email} maxlength={254} />
        {/snippet}
      </Field>
      <Field label="Name" hint="Optional. Pre-fills their application.">
        {#snippet children({ id, describedby })}
          <Input {id} aria-describedby={describedby} bind:value={displayName} maxlength={100} />
        {/snippet}
      </Field>
      <Field label="Reserved handle" error={fieldErrors.handle} hint="Optional. Nobody else can take it while the invite is waiting.">
        {#snippet children({ id, describedby, invalid })}
          <Input {id} class="font-mono" autocomplete="off" aria-describedby={describedby} {invalid} bind:value={handle} maxlength={30} />
        {/snippet}
      </Field>
      <Field label="Link valid for">
        {#snippet children({ id })}
          <Select {id} bind:value={ttlDays}>
            <option value="7">7 days</option>
            <option value="14">14 days</option>
            <option value="30">30 days</option>
            <option value="60">60 days</option>
          </Select>
        {/snippet}
      </Field>
      <div class="md:col-span-2">
        <Button type="submit" disabled={submitting}>{submitting ? 'Creating…' : 'Create invite'}</Button>
      </div>
    </form>
  </CardContent>
</Card>

{#if data.invites.length === 0}
  <EmptyState title="No invites yet" description="Invites you create will show up here with their status." />
{:else}
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Invitee</TableHead>
        <TableHead>Handle</TableHead>
        <TableHead>Status</TableHead>
        <TableHead>Expires</TableHead>
        <TableHead><span class="sr-only">Actions</span></TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.invites as invite (invite.id)}
        <TableRow data-testid="invite-row">
          <TableCell>
            <div class="font-medium text-foreground">{invite.display_name || invite.email}</div>
            {#if invite.display_name}<div class="text-xs text-muted-foreground">{invite.email}</div>{/if}
          </TableCell>
          <TableCell class="font-mono text-xs text-muted-foreground">
            {invite.requested_handle ? `@${invite.requested_handle}` : '—'}
          </TableCell>
          <TableCell><Badge variant={statusVariant[invite.status]}>{statusLabel[invite.status]}</Badge></TableCell>
          <TableCell class="text-muted-foreground">{formatDate(invite.expires_at)}</TableCell>
          <TableCell class="text-right">
            {#if invite.status !== 'claimed'}
              <div class="flex justify-end gap-2">
                <Button variant="outline" size="sm" onclick={() => askResend(invite)}>Resend</Button>
                {#if invite.status === 'live'}
                  <Button variant="ghost" size="sm" onclick={() => askRevoke(invite)}>Revoke</Button>
                {/if}
              </div>
            {/if}
          </TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
