<script lang="ts">
	import { page } from '$app/stores';
	import { apiFetch } from '$lib/adapters/api/client';
	import { onMount } from 'svelte';

	let status = $state<'working' | 'success' | 'error'>('working');

	onMount(async () => {
		const token = $page.url.searchParams.get('token');
		if (!token) {
			status = 'error';
			return;
		}
		try {
			await apiFetch('/api/v1/auth/verify-email', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ token })
			});
			status = 'success';
		} catch {
			status = 'error';
		}
	});
</script>

<svelte:head><title>Verify email — Artiv</title></svelte:head>

<main class="mx-auto max-w-xl px-6 py-24 text-center">
	<h1 class="font-display text-4xl">Email verification</h1>
	{#if status === 'working'}
		<p class="mt-5 text-muted-foreground">Verifying your email…</p>
	{:else if status === 'success'}
		<p class="mt-5">Your email is verified. You can continue to Artiv.</p>
		<a class="mt-8 inline-block underline" href="/">Continue</a>
	{:else}
		<p class="mt-5 text-muted-foreground">This verification link is invalid or expired.</p>
	{/if}
</main>
