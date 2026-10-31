<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService } from '$lib/application/auth';
	import { inviteService } from '$lib/application/invites';
	import PageSkeleton from '$lib/components/PageSkeleton.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			// Keep the query string: an invite link is /apply?invite=…, and losing it here would drop
			// the invitee onto the plain form after signing in.
			const here = $page.url.pathname + $page.url.search;
			let emailParam = '';
			const token = $page.url.searchParams.get('invite');
			if (token) {
				try {
					// Pre-fill the invited address so they sign in with the right account.
					emailParam = `&email=${encodeURIComponent((await inviteService.preview(token)).email)}`;
				} catch {
					// An unknown or broken invite shouldn't stop sign-in; the page explains it afterwards.
				}
			}
			await goto(`/login?intent=artist&return_to=${encodeURIComponent(here)}${emailParam}`);
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
