<svelte:head>
  <title>Login — mq admin</title>
</svelte:head>

<script lang="ts">
  import { page } from '$app/stores';
  import { googleLoginUrl } from '$lib/core/domain/auth';

  const loginUrl = googleLoginUrl();
  const accessDenied = $derived($page.url.searchParams.get('error') === 'access_denied');
</script>

<h1>Admin login</h1>
<p>Sign in with a Google account that has admin access.</p>

{#if accessDenied}
  <p role="alert">Access denied. Your account does not have admin privileges.</p>
{/if}

<a class="google-btn" href={loginUrl}>Sign in with Google</a>

<style>
  .google-btn {
    display: inline-block;
    padding: 0.75rem 1.25rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    text-decoration: none;
    color: inherit;
  }
</style>
