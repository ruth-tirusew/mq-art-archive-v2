<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService, currentUser, authLoading } from '$lib/application/auth';

	let { children }: { children: import('svelte').Snippet } = $props();
	let ready = $state(false);

	const studioNav = $derived(
		$currentUser?.role === 'contributor'
			? [{ href: '/studio/wiki', label: 'Wiki' }]
			: [
					{ href: '/studio', label: 'Overview' },
					{ href: '/studio/profile', label: 'Profile' },
					{ href: '/studio/posts', label: 'Posts' },
					{ href: '/studio/wiki', label: 'Wiki' }
				]
	);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			const returnTo = $page.url.pathname;
			await goto(`/login?return_to=${encodeURIComponent(returnTo)}`);
			return;
		}
		if (user.role === 'contributor' && $page.url.pathname === '/studio') {
			await goto('/studio/wiki');
			return;
		}
		ready = true;
	});

	function isActive(href: string) {
		return $page.url.pathname === href;
	}
</script>

<svelte:head>
	<title>Studio — Artiv</title>
</svelte:head>

{#if $authLoading || !ready}
	<p class="mx-auto max-w-3xl px-6 py-24 text-muted-foreground">Loading studio…</p>
{:else if $currentUser?.role !== 'artist' && $currentUser?.role !== 'contributor'}
	<section class="mx-auto max-w-2xl px-6 py-24 md:px-10">
		<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Studio access</p>
		<h1 class="mt-4 font-display text-3xl text-foreground">Artist approval required</h1>
		<p class="mt-4 text-sm leading-relaxed text-muted-foreground">
			Your account is signed in, but you need an approved artist application before you can manage a
			profile here.
		</p>
		<div class="mt-8 flex flex-wrap gap-3">
			<a
				href="/apply"
				class="inline-flex rounded-sm border border-border bg-card px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-foreground transition hover:border-foreground/30"
			>
				Apply as artist
			</a>
			<a href="/apply/status" class="inline-flex px-2 py-3 text-sm text-accent underline underline-offset-4">
				Check application status
			</a>
		</div>
	</section>
{:else}
	<div class="border-b border-border/60 bg-card/20">
		<div class="mx-auto flex max-w-[1600px] flex-wrap items-center gap-4 px-6 py-4 md:px-10">
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
				{$currentUser?.role === 'contributor' ? 'Contributor studio' : 'Artist studio'}
			</p>
			<nav class="flex flex-wrap gap-4 font-mono text-[11px] uppercase tracking-[0.2em]">
				{#each studioNav as item}
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
