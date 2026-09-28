<script lang="ts">
	import { goto } from '$app/navigation';
	import { authService, currentUser } from '$lib/application/auth';

	let open = $state(false);
	let root: HTMLDivElement | undefined = $state();

	const initial = $derived(
		($currentUser?.display_name?.trim()?.[0] ?? $currentUser?.email?.[0] ?? '?').toUpperCase()
	);
	const label = $derived($currentUser?.display_name?.trim() || $currentUser?.email || '');

	function toggle() {
		open = !open;
	}

	function close() {
		open = false;
	}

	function handleWindowClick(event: MouseEvent) {
		if (open && root && !root.contains(event.target as Node)) close();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') close();
	}

	async function logout() {
		close();
		await authService.logout();
		await goto('/');
	}
</script>

<svelte:window onclick={handleWindowClick} onkeydown={handleKeydown} />

<div class="relative normal-case" bind:this={root}>
	<button
		type="button"
		data-testid="web-profile-dropdown-trigger"
		onclick={toggle}
		aria-expanded={open}
		aria-haspopup="menu"
		class="flex size-8 items-center justify-center overflow-hidden rounded-full border border-border bg-card font-mono text-[11px] uppercase text-foreground transition hover:border-foreground/40"
		title={label}
	>
		{#if $currentUser?.avatar_url}
			<img src={$currentUser.avatar_url} alt="" class="size-full object-cover" />
		{:else}
			{initial}
		{/if}
	</button>

	{#if open}
		<div
			role="menu"
			data-testid="web-profile-dropdown-menu"
			class="absolute right-0 top-full mt-2 w-56 rounded-sm border border-border bg-card py-2 shadow-lg"
		>
			<p class="truncate border-b border-border/60 px-4 pb-2 text-xs text-muted-foreground" title={$currentUser?.email}>
				{label}
			</p>
			<a
				href="/account"
				role="menuitem"
				onclick={close}
				class="block px-4 py-2 text-xs text-foreground transition hover:bg-muted"
			>
				Profile settings
			</a>
			<a
				href="/account/activity"
				role="menuitem"
				onclick={close}
				class="block px-4 py-2 text-xs text-foreground transition hover:bg-muted"
			>
				My activity
			</a>
			<a
				href="/wiki/saved"
				role="menuitem"
				onclick={close}
				class="block px-4 py-2 text-xs text-foreground transition hover:bg-muted"
			>
				Saved articles
			</a>
			<div class="mt-1 border-t border-border/60 pt-1">
				<button
					type="button"
					role="menuitem"
					onclick={logout}
					class="block w-full px-4 py-2 text-left text-xs text-destructive transition hover:bg-muted"
				>
					Sign out
				</button>
			</div>
		</div>
	{/if}
</div>
