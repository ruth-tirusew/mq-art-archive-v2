<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { favoritesService } from '$lib/application/favorites';
	import { ApiError } from '$lib/adapters/api/client';

	let { articleId }: { articleId: string } = $props();

	let favorited = $state(false);
	let count = $state(0);
	let busy = $state(false);

	onMount(async () => {
		try {
			const status = await favoritesService.getStatus(articleId);
			favorited = status.favorited;
			count = status.count;
		} catch {
			// leave at defaults — this is a nice-to-have, not worth an error state
		}
	});

	async function toggle() {
		if (busy) return;
		busy = true;
		try {
			const status = await favoritesService.toggle(articleId);
			favorited = status.favorited;
			count = status.count;
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				await goto(`/login?return_to=${encodeURIComponent($page.url.pathname)}`);
			}
		} finally {
			busy = false;
		}
	}
</script>

<button
	type="button"
	data-testid="web-wiki-favorite-button"
	onclick={toggle}
	disabled={busy}
	aria-pressed={favorited}
	class="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 font-mono text-[10px] uppercase tracking-[0.2em] transition disabled:opacity-50 {favorited
		? 'border-accent bg-accent/10 text-accent'
		: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
>
	<span aria-hidden="true">{favorited ? '★' : '☆'}</span>
	{favorited ? 'Saved' : 'Save'}
	{#if count > 0}<span class="opacity-70">· {count}</span>{/if}
</button>
