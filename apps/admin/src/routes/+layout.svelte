<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { authService, currentUser } from '$lib/application/auth';

  let { children } = $props();

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
</script>

<header>
  <nav>
    <a href="/">mq admin</a>
    <a href="/applications">Applications</a>
    {#if $currentUser}
      <span>{$currentUser.email}</span>
      <button type="button" onclick={logout}>Log out</button>
    {:else}
      <a href="/login">Login</a>
    {/if}
  </nav>
</header>

<main>
  {@render children()}
</main>

<style>
  header {
    border-bottom: 1px solid #ddd;
    margin-bottom: 1.5rem;
    padding: 1rem 0;
  }

  nav {
    display: flex;
    gap: 1rem;
    max-width: 960px;
    margin: 0 auto;
    padding: 0 1rem;
  }

  main {
    max-width: 960px;
    margin: 0 auto;
    padding: 0 1rem 2rem;
  }
</style>
