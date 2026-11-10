<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { authService } from '$lib/application/auth';
  import { isStaffRole } from '$lib/utils/adminGate';

  onMount(async () => {
    // The API sends us back here with ?error=… when Google sign-in did not end in a session, most often
    // account_exists (that Google email already has a password account, which is never linked automatically).
    // Hand the reason to the login page so it can explain, instead of showing a login form as if nothing happened.
    const oauthError = $page.url.searchParams.get('error');
    if (oauthError) {
      await goto(`/login?error=${encodeURIComponent(oauthError)}`);
      return;
    }

    const user = await authService.load();
    if (user && isStaffRole(user.role)) {
      await goto('/applications');
    } else if (user) {
      await goto('/login?error=access_denied');
    } else {
      // Google said yes but there is no session: the cookie did not stick, or the sign-in expired.
      await goto('/login?error=google_failed');
    }
  });
</script>

<p>Signing you in…</p>
