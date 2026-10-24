<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService, currentUser } from '$lib/application/auth';
	import PageSkeleton from '$lib/components/PageSkeleton.svelte';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);

	const accountNav = [
		{ href: '/account', label: 'Profile' },
		{ href: '/account/activity', label: 'Activity' }
	];

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			const returnTo = $page.url.pathname;
			await goto(`/login?return_to=${encodeURIComponent(returnTo)}`);
			return;
		}
		ready = true;
	});

	function isActive(href: string) {
		return $page.url.pathname === href;
	}
</script>

<svelte:head>
	<title>Account — Artiv</title>
</svelte:head>

{#if !ready}
	<PageSkeleton rows={3} label="Loading account" class="mx-auto max-w-3xl px-6 py-16 md:px-10 md:py-20" />
{:else}
	<div class="border-b border-border/60 bg-card/20">
		<div class="mx-auto flex max-w-3xl flex-wrap items-center gap-4 px-6 py-4 md:px-10">
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Account</p>
			<nav class="flex flex-wrap gap-4 font-mono text-[11px] uppercase tracking-[0.2em]">
				{#each accountNav as item}
					<a
						href={item.href}
						class="transition {isActive(item.href)
							? 'text-foreground'
							: 'text-muted-foreground hover:text-foreground'}"
					>
						{item.label}
					</a>
				{/each}
			</nav>
		</div>
	</div>
	{@render children()}
{/if}
