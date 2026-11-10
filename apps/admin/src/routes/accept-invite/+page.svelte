<script lang="ts">
  import { goto } from '$app/navigation';
  import { ApiError } from '$lib/adapters/api/client';
  import { currentUser } from '$lib/application/auth';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { staffInvitesService } from '$lib/composition';
  import { passwordProblem } from '$lib/core/domain/staffInvite';
  import { Field, PasswordInput } from '$ui';

  let { data } = $props();

  let password = $state('');
  let passwordError = $state<string | undefined>();
  let submitting = $state(false);
  let error = $state('');

  const invite = $derived(data.state);

  const messages = {
    invalid: ['This invitation link isn’t valid', 'It may have been copied incompletely. Ask an admin to send you a new invitation.'],
    used: ['This invitation has already been used', 'If it was you, sign in with your email and the password you chose.'],
    revoked: ['This invitation was cancelled', 'Ask an admin to send you a new one.'],
    expired: ['This invitation has expired', 'Ask an admin to send you a new one.'],
    error: ['We couldn’t check this invitation', 'The server didn’t answer. Try again in a moment.']
  } as const;

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = '';
    passwordError = passwordProblem(password) ?? undefined;
    if (passwordError) return;
    submitting = true;
    try {
      const user = await staffInvitesService.accept(data.token, password);
      currentUser.set(user);
      await goto('/applications');
    } catch (cause) {
      if (cause instanceof ApiError && cause.status === 409) {
        error = 'This invitation can no longer be used. It may have been used, cancelled or expired. Ask an admin for a new one.';
      } else {
        error = cause instanceof Error ? cause.message : 'Something went wrong. Please try again.';
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Join as a moderator — Artiv admin</title>
  <!-- The address holds a one-time token; keep it out of search results and referrers. -->
  <meta name="robots" content="noindex" />
  <meta name="referrer" content="no-referrer" />
</svelte:head>

<div class="flex min-h-screen items-center justify-center px-6 py-12">
  <Card class="w-full max-w-md">
    {#if invite.kind === 'ready'}
      <CardHeader>
        <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">Artiv admin</p>
        <CardTitle>Join as a moderator</CardTitle>
        <CardDescription>
          Choose a password for <strong class="text-foreground" data-testid="invited-email">{invite.preview.email}</strong> to finish setting up your account.
        </CardDescription>
      </CardHeader>
      <CardContent class="space-y-4">
        {#if error}<Alert variant="destructive">{error}</Alert>{/if}
        <form class="space-y-4" onsubmit={submit} novalidate>
          <Field label="Password" required error={passwordError} hint="At least 8 characters.">
            {#snippet children({ id, describedby, invalid })}
              <PasswordInput {id} {invalid} aria-describedby={describedby} autocomplete="new-password" bind:value={password} />
            {/snippet}
          </Field>
          <Button type="submit" class="w-full" disabled={submitting}>{submitting ? 'Please wait…' : 'Create my account'}</Button>
        </form>
        <p class="text-xs text-muted-foreground">
          As a moderator you can review applications and moderate content. You can’t manage users or change system settings.
        </p>
      </CardContent>
    {:else}
      <CardHeader>
        <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">Artiv admin</p>
        <div data-testid="invite-problem"><CardTitle>{messages[invite.kind][0]}</CardTitle></div>
        <CardDescription>{messages[invite.kind][1]}</CardDescription>
      </CardHeader>
      <CardContent>
        <Button href="/login" variant="outline" class="w-full">Go to sign in</Button>
      </CardContent>
    {/if}
  </Card>
</div>
