<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { authService } from '$lib/application/auth';

  onMount(async () => {
    const user = await authService.load();
    if (user?.role === 'admin') {
      await goto('/applications');
    } else if (user) {
      await goto('/login?error=access_denied');
    } else {
      await goto('/login');
    }
  });
</script>

<p>Signing you in…</p>
