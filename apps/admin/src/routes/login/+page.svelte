<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { googleLoginUrl } from '$lib/core/domain/auth';
  import { authService } from '$lib/application/auth';
  import { ApiError } from '$lib/adapters/api/client';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Separator } from '$lib/components/ui/separator';

  type Mode = 'signin' | 'signup';

  const loginUrl = $derived(googleLoginUrl($page.url.origin));
  const accessDenied = $derived($page.url.searchParams.get('error') === 'access_denied');

  let mode = $state<Mode>('signin');
  let email = $state('');
  let password = $state('');
  let submitting = $state(false);
  let error = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    error = '';
    submitting = true;
    try {
      const user =
        mode === 'signup'
          ? await authService.register(email, password)
          : await authService.login(email, password);

      if (user.role === 'admin') {
        await goto('/applications');
      } else {
        await authService.logout();
        await goto('/login?error=access_denied');
      }
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 409) {
          error = 'An account with this email already exists.';
        } else if (err.status === 401) {
          error = 'Invalid email or password.';
        } else {
          error = err.message;
        }
      } else {
        error = 'Something went wrong. Please try again.';
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Login — mq admin</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center px-6 py-12">
  <Card class="w-full max-w-md">
    <CardHeader>
      <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">mq admin</p>
      <CardTitle>{mode === 'signup' ? 'Create account' : 'Sign in'}</CardTitle>
      <CardDescription>
        Use email and password or Google. Admin privileges are required to continue.
      </CardDescription>
    </CardHeader>
    <CardContent class="space-y-4">
      {#if accessDenied}
        <Alert variant="destructive">Access denied. Your account does not have admin privileges.</Alert>
      {/if}
      {#if error}
        <Alert variant="destructive">{error}</Alert>
      {/if}

      <form class="space-y-4" data-testid="admin-login-form" onsubmit={handleSubmit}>
        <div class="space-y-2">
          <Label for="email">Email</Label>
          <input
            id="email"
            type="email"
            required
            autocomplete="email"
            bind:value={email}
            class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
          />
        </div>
        <div class="space-y-2">
          <Label for="password">Password</Label>
          <input
            id="password"
            type="password"
            required
            minlength="8"
            autocomplete={mode === 'signup' ? 'new-password' : 'current-password'}
            bind:value={password}
            class="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50"
          />
        </div>
        <Button type="submit" class="w-full" disabled={submitting}>
          {submitting ? 'Please wait…' : mode === 'signup' ? 'Create account' : 'Sign in'}
        </Button>
      </form>
      <div class="flex items-center gap-3">
        <Separator class="flex-1" />
        <span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">or</span>
        <Separator class="flex-1" />
      </div>

      <Button href={loginUrl} variant="outline" class="w-full">Continue with Google</Button>
    </CardContent>
  </Card>
</div>
