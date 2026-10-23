<script lang="ts">
  import { page } from '$app/stores';
  import { Button, ErrorState } from '$ui';

  // Only surface the server's message for client errors; 5xx text is not user-facing.
  const message = $derived($page.status < 500 ? $page.error?.message : undefined);
</script>

<svelte:head>
  <title>{$page.status} — Artiv admin</title>
</svelte:head>

<div class="mx-auto flex max-w-2xl flex-col gap-6 py-16">
  <ErrorState status={$page.status} {message} />
  <div><Button variant="link" href="/">Back to overview →</Button></div>
</div>
