<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authService, currentUser, authLoading } from '$lib/application/auth';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			await goto('/login?return_to=/apply');
			return;
		}
		if (user.role === 'artist') {
			await goto('/studio');
			return;
		}
		ready = true;
	});
</script>

{#if $authLoading || !ready}
	<p class="mx-auto max-w-2xl px-6 py-24 text-muted-foreground">Loading…</p>
{:else}
	{@render children()}
{/if}
