<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService } from '$lib/application/auth';
	import PageSkeleton from '$lib/components/PageSkeleton.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			await goto(`/login?intent=artist&return_to=${encodeURIComponent($page.url.pathname)}`);
			return;
		}
		if (user.role === 'artist') {
			await goto('/studio');
			return;
		}
		if (user.role === 'institution' && $page.url.pathname !== '/apply/status') {
			await goto('/apply/status');
			return;
		}
		ready = true;
	});
</script>

{#if !ready}
	<PageSkeleton rows={4} label="Loading" />
{:else}
	{@render children()}
{/if}
