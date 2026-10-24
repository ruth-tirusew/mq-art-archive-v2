<script lang="ts">
	import { onMount } from 'svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService, postLoginPath } from '$lib/application/auth';
	import { safeReturnPath } from '$lib/core/domain/auth';

	onMount(async () => {
		const returnTo = safeReturnPath($page.url.searchParams.get('return_to'));
		const oauthError = $page.url.searchParams.get('error');

		// The API sends us back here with ?error=account_exists when that Google email already
		// has an Artiv account: explain it on the sign-in page rather than failing silently.
		if (oauthError) {
			await goto(`/login?error=${encodeURIComponent(oauthError)}&return_to=${encodeURIComponent(returnTo)}`);
			return;
		}

		const user = await authService.load();
		if (!user) {
			await goto(`/login?return_to=${encodeURIComponent(returnTo)}`);
			return;
		}

		await goto(postLoginPath(user, returnTo));
	});
</script>

<div class="mx-auto max-w-lg space-y-4 px-6 py-24" role="status" aria-live="polite">
	<span class="sr-only">Signing you in</span>
	<Skeleton class="mx-auto h-8 w-48" />
	<Skeleton class="mx-auto h-4 w-64" />
</div>
