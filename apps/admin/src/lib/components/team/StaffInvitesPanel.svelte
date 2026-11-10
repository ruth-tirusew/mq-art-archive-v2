<script lang="ts">
  import { onMount } from 'svelte';
  import { ConfirmDialog, Field, Input, createConfirm, toast } from '$ui';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
  import { staffInvitesService } from '$lib/composition';
  import {
    canResendStaffInvite,
    canRevokeStaffInvite,
    staffInviteEmailProblem,
    staffInviteStatusLabel,
    type StaffInvite,
    type StaffInviteCreated,
    type StaffInviteStatus
  } from '$lib/core/domain/staffInvite';

  const confirm = createConfirm();

  let email = $state('');
  let emailError = $state<string | undefined>();
  let submitting = $state(false);
  let errorMessage = $state('');
  /** The link is returned once, so it is held here until the admin dismisses it. */
  let created = $state<StaffInviteCreated | null>(null);

  let invites = $state<StaffInvite[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let loadingMore = $state(false);
  let loadError = $state('');

  const hasMore = $derived(invites.length < total);

  const statusVariant: Record<StaffInviteStatus, 'warning' | 'success' | 'muted' | 'destructive'> = {
    live: 'warning',
    claimed: 'success',
    expired: 'muted',
    revoked: 'destructive'
  };

  const formatDate = (iso: string) => new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short', year: 'numeric' });

  async function load(more = false) {
    loadError = '';
    if (more) loadingMore = true;
    else loading = true;
    try {
      const page = await staffInvitesService.list(more ? invites.length : 0);
      invites = more ? [...invites, ...page.items] : page.items;
      total = page.total;
    } catch (cause) {
      loadError = cause instanceof Error ? cause.message : 'Could not load invites';
    } finally {
      loading = false;
      loadingMore = false;
    }
  }

  onMount(() => void load());

  async function send(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = '';
    emailError = staffInviteEmailProblem(email) ?? undefined;
    if (emailError) return;
    submitting = true;
    try {
      created = await staffInvitesService.create(email);
      email = '';
      toast.success('Invite created.');
      await load();
    } catch (cause) {
      // The API says why, e.g. "x@y already has an Artiv account. Change their role on the Users page instead".
      errorMessage = cause instanceof Error ? cause.message : 'Could not create the invite';
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

  function askResend(invite: StaffInvite) {
    confirm.ask({
      title: `Send ${invite.email} a new link?`,
      description: 'The old link stops working immediately.',
      confirmLabel: 'Send new link',
      run: async () => {
        try {
          created = await staffInvitesService.resend(invite.id);
          await load();
          toast.success('New link created.');
        } catch (cause) {
          errorMessage = cause instanceof Error ? cause.message : 'Could not resend the invite';
        }
      }
    });
  }

  function askRevoke(invite: StaffInvite) {
    confirm.ask({
      title: `Cancel the invite for ${invite.email}?`,
      description: 'The link stops working. No account is created.',
      confirmLabel: 'Cancel invite',
      destructive: true,
      run: async () => {
        try {
          await staffInvitesService.revoke(invite.id);
          await load();
          toast.success('Invite cancelled.');
        } catch (cause) {
          errorMessage = cause instanceof Error ? cause.message : 'Could not cancel the invite';
        }
      }
    });
  }
</script>

<section class="mb-10 space-y-6" aria-labelledby="staff-invites-heading" data-testid="staff-invites">
  <div>
    <h2 id="staff-invites-heading" class="text-lg font-semibold text-foreground">Invite a moderator</h2>
    <p class="mt-1 max-w-2xl text-sm text-muted-foreground">
      A moderator can review applications, approve wiki submissions, moderate artworks and artists, and manage events and content.
      They can't manage users or change system settings. To make someone who already has an account a moderator, change their role in the table below.
    </p>
  </div>

  {#if errorMessage}
    <Alert variant="destructive">{errorMessage}</Alert>
  {/if}

  {#if created}
    <div data-testid="staff-invite-link">
      <Card class="border-accent/50">
        <CardHeader>
          <CardTitle>Invite link for {created.invite.email}</CardTitle>
        </CardHeader>
        <CardContent class="space-y-3">
          <p class="text-sm text-muted-foreground">
            This is the only time the link is shown. We also emailed it, but if mail isn't set up, copy it now and send it yourself. It works once.
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

  <form class="flex flex-col gap-3 sm:flex-row sm:items-end" onsubmit={send} novalidate>
    <div class="flex-1">
      <Field label="Email" required error={emailError} hint="Where the link is sent. The account they create will use this address.">
        {#snippet children({ id, describedby, invalid })}
          <Input {id} type="email" autocomplete="off" aria-describedby={describedby} {invalid} bind:value={email} maxlength={254} />
        {/snippet}
      </Field>
    </div>
    <Button type="submit" disabled={submitting}>{submitting ? 'Sending…' : 'Send invite'}</Button>
  </form>

  {#if loading}
    <div class="space-y-2" role="status" aria-live="polite">
      <span class="sr-only">Loading invites</span>
      <Skeleton class="h-10 w-full" />
      <Skeleton class="h-10 w-full" />
    </div>
  {:else if loadError}
    <Alert variant="destructive">{loadError} <Button variant="secondary" size="sm" onclick={() => load()}>Try again</Button></Alert>
  {:else if invites.length === 0}
    <EmptyState title="No moderator invites yet" description="Invites you send will show up here with their status." />
  {:else}
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Invitee</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Expires</TableHead>
          <TableHead><span class="sr-only">Actions</span></TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {#each invites as invite (invite.id)}
          <TableRow data-testid="staff-invite-row">
            <TableCell class="font-medium text-foreground">{invite.email}</TableCell>
            <TableCell><Badge variant={statusVariant[invite.status]}>{staffInviteStatusLabel[invite.status]}</Badge></TableCell>
            <TableCell class="text-muted-foreground">{invite.status === 'claimed' ? '—' : formatDate(invite.expires_at)}</TableCell>
            <TableCell class="text-right">
              {#if canResendStaffInvite(invite)}
                <div class="flex justify-end gap-2">
                  <Button variant="outline" size="sm" onclick={() => askResend(invite)}>Resend</Button>
                  {#if canRevokeStaffInvite(invite)}
                    <Button variant="ghost" size="sm" onclick={() => askRevoke(invite)}>Cancel</Button>
                  {/if}
                </div>
              {/if}
            </TableCell>
          </TableRow>
        {/each}
      </TableBody>
    </Table>
    <p class="text-xs text-muted-foreground">Showing {invites.length} of {total}.</p>
    {#if hasMore}
      <Button variant="secondary" size="sm" disabled={loadingMore} onclick={() => load(true)}>{loadingMore ? 'Loading…' : 'Load more'}</Button>
    {/if}
  {/if}
</section>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
