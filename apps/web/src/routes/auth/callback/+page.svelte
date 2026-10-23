<script lang="ts">
	import { onMount } from 'svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService, postLoginPath } from '$lib/application/auth';

	onMount(async () => {
		const user = await authService.load();
		const returnTo = $page.url.searchParams.get('return_to') ?? '/studio';

		if (!user) {
			await goto('/login');
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
