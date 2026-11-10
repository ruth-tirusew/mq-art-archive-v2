<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService } from '$lib/application/auth';
	import { inviteService } from '$lib/composition';
	import { isErrorKind } from '$lib/core/domain/errors';
	import InviteLanding from '$lib/components/InviteLanding.svelte';
	import InviteProblem, { type InviteProblemKind } from '$lib/components/InviteProblem.svelte';
	import PageSkeleton from '$lib/components/PageSkeleton.svelte';
	import type { InvitePreview } from '$lib/core/domain/invite';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);
	// Set for a signed-out visitor holding a live invite: they register right here, no sign-in page.
	let landing = $state<InvitePreview | null>(null);
	const inviteToken = $derived($page.url.searchParams.get('invite'));
	// A signed-out visitor with an invite link that can't be used sees WHY, not a silent redirect.
	let problem = $state<InviteProblemKind | null>(null);
	let retrying = $state(false);

	async function checkInvite(token: string) {
		problem = null;
		try {
			const preview = await inviteService.preview(token);
			if (preview.status === 'live') {
				// Someone with no account can join straight from the link: show the registration
				// panel here instead of sending them to a sign-in page they don't need.
				landing = preview;
			} else {
				problem = preview.status;
			}
		} catch (e) {
			problem = isErrorKind(e, 'not_found') ? 'invalid' : 'unreachable';
		}
	}

	async function retry() {
		if (!inviteToken) return;
		retrying = true;
		await checkInvite(inviteToken);
		retrying = false;
	}

	const here = $derived($page.url.pathname + $page.url.search);
	const signInHref = $derived(`/login?intent=artist&return_to=${encodeURIComponent(here)}`);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			const token = $page.url.searchParams.get('invite');
			if (token) {
				// Never fall through to the sign-in page without saying why: an invite link that
				// can't be used must explain itself.
				await checkInvite(token);
				return;
			}
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

{#if problem}
	<InviteProblem kind={problem} {signInHref} applyHref="/login?intent=artist&return_to=%2Fapply" onretry={retry} {retrying} />
{:else if landing && inviteToken}
	<InviteLanding
		invite={landing}
		token={inviteToken}
		onregistered={() => {
			landing = null;
			ready = true;
		}}
	/>
{:else if !ready}
	<PageSkeleton rows={4} label="Loading" />
{:else}
	{@render children()}
{/if}
