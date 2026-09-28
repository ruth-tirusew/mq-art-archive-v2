<script lang="ts">
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button';

  const status = $derived($page.status);
  const message = $derived($page.error?.message ?? 'Something went wrong.');
  const isUnreachable = $derived(status === 503);
</script>

<svelte:head>
  <title>{status} — mq admin</title>
</svelte:head>

<div class="mx-auto flex min-h-[70vh] max-w-md flex-col items-center justify-center px-6 text-center">
  <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
    Error {status}
  </p>
  <h1 class="mt-4 text-2xl font-medium text-foreground">
    {isUnreachable ? "Can't reach the API" : 'Something went wrong'}
  </h1>
  <p class="mt-3 text-sm text-muted-foreground">{message}</p>
  <div class="mt-8 flex gap-3">
    <Button onclick={() => location.reload()}>Retry</Button>
    <Button variant="outline" href="/">Back to overview</Button>
  </div>
</div>
