<script lang="ts">
	import { page } from '$app/state';
	import { Button, ErrorState } from '$ui';

	// Only surface the server's message for client errors; 5xx text is not user-facing.
	const message = $derived(page.status < 500 ? page.error?.message : undefined);
</script>

<svelte:head>
	<title>{page.status === 404 ? 'Not found' : 'Something went wrong'} — Artiv</title>
</svelte:head>

<div class="container-page py-24 md:py-32">
	<div class="mx-auto flex max-w-2xl flex-col gap-6">
		<ErrorState status={page.status} {message} />
		<div><Button variant="link" href="/">Back home →</Button></div>
	</div>
</div>
