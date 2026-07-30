<script lang="ts">
	import { onMount } from 'svelte';
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

<p class="mx-auto max-w-lg px-6 py-24 text-center text-muted-foreground">Signing you in…</p>
