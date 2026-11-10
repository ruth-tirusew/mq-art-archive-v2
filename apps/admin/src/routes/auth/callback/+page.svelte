<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { authService } from '$lib/application/auth';
  import { isStaffRole } from '$lib/utils/adminGate';

  onMount(async () => {
    const user = await authService.load();
    if (user && isStaffRole(user.role)) {
      await goto('/applications');
    } else if (user) {
      await goto('/login?error=access_denied');
    } else {
      await goto('/login');
    }
  });
</script>

<p>Signing you in…</p>
