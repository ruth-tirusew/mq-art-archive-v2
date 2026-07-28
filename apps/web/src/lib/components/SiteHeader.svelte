<script lang="ts">
	import { page } from '$app/stores';
	import SearchTrigger from './GlobalSearch.svelte';
	import { authService, currentUser, authLoading } from '$lib/application/auth';

	const primaryNav = [
		{ href: '/artists', label: 'Artists' },
		{ href: '/archive', label: 'Archive' },
		{ href: '/events', label: 'Events' },
		{ href: '/wiki', label: 'Wiki' }
	];

	const secondaryNav = [
		{ href: '/about', label: 'About' },
		{ href: '/portfolio', label: 'Portfolio' }
	];

	const navLinkClass =
		'relative text-foreground/85 transition after:absolute after:-bottom-1 after:left-0 after:h-px after:w-full after:origin-left after:scale-x-0 after:bg-foreground after:transition-transform after:duration-300 hover:text-foreground hover:after:scale-x-100';

	const mobileLinkClass =
		'block py-2.5 text-sm text-foreground/85 transition hover:text-foreground';

	let mobileOpen = $state(false);

	function isActive(href: string) {
		return $page.url.pathname === href || $page.url.pathname.startsWith(href + '/');
	}

	function toggleMenu() {
		mobileOpen = !mobileOpen;
	}

	function closeMenu() {
		mobileOpen = false;
	}

	$effect(() => {
		void $page.url.pathname;
		mobileOpen = false;
	});

	$effect(() => {
		if (typeof document === 'undefined') return;

		document.body.style.overflow = mobileOpen ? 'hidden' : '';

		return () => {
			document.body.style.overflow = '';
		};
	});

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') closeMenu();
	}

	async function logout() {
		await authService.logout();
		closeMenu();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<header
	data-testid="web-site-header"
	class="sticky top-0 z-40 border-b border-border/60 bg-background/85 backdrop-blur"
>
	<div class="mx-auto flex max-w-[1600px] items-center justify-between gap-3 px-6 py-4 md:gap-4 md:px-10">
		<a href="/" class="group flex min-w-0 items-baseline gap-2" onclick={closeMenu}>
			<span
				class="font-display text-2xl font-medium tracking-tight text-foreground transition group-hover:text-accent"
			>
				artiv.
			</span>
		</a>

		<div class="flex items-center gap-2 sm:gap-4 md:gap-6">
			<SearchTrigger />

			<button
				type="button"
				class="inline-flex size-9 items-center justify-center rounded-sm border border-border text-foreground transition hover:border-foreground/20 hover:bg-muted lg:hidden"
				aria-expanded={mobileOpen}
				aria-controls="mobile-nav"
				aria-label={mobileOpen ? 'Close menu' : 'Open menu'}
				onclick={toggleMenu}
			>
				{#if mobileOpen}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
						class="size-5"
						aria-hidden="true"
					>
						<path stroke-linecap="round" d="M6 6l12 12M18 6L6 18" />
					</svg>
				{:else}
					<svg
						xmlns="http://www.w3.org/2000/svg"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="1.5"
						class="size-5"
						aria-hidden="true"
					>
						<path stroke-linecap="round" d="M4 7h16M4 12h16M4 17h16" />
					</svg>
				{/if}
			</button>

			<nav class="hidden items-center gap-x-4 gap-y-2 font-medium text-[12px] uppercase lg:flex">
				{#each primaryNav as n}
					<a
						href={n.href}
						class="{navLinkClass} {isActive(n.href) ? 'text-foreground after:scale-x-100' : ''}"
					>
						{n.label}
					</a>
				{/each}

				<span class="h-3 w-px bg-border" aria-hidden="true"></span>

				{#each secondaryNav as n}
					<a
						href={n.href}
						class="text-muted-foreground transition hover:text-foreground {isActive(n.href)
							? 'text-foreground'
							: ''}"
					>
						{n.label}
					</a>
				{/each}

				<span class="h-3 w-px bg-border" aria-hidden="true"></span>

				{#if $currentUser?.role === 'artist'}
					<a
						href="/studio"
						class="{navLinkClass} {isActive('/studio') ? 'text-foreground after:scale-x-100' : ''}"
					>
						Studio
					</a>
				{/if}

				{#if $authLoading}
					<span class="text-muted-foreground">…</span>
				{:else if $currentUser}
					<span class="hidden max-w-[10rem] truncate text-muted-foreground xl:inline" title={$currentUser.email}>
						{$currentUser.email}
					</span>
					<button
						type="button"
						class="text-muted-foreground transition hover:text-foreground"
						onclick={logout}
					>
						Sign out
					</button>
				{:else}
					<a href="/login" class="text-muted-foreground transition hover:text-foreground">Sign in</a>
				{/if}
			</nav>
		</div>
	</div>

	{#if mobileOpen}
		<div class="fixed inset-0 top-[65px] z-30 bg-foreground/20 lg:hidden" aria-hidden="true" onclick={closeMenu}>
		</div>

		<nav
			id="mobile-nav"
			class="relative z-40 border-t border-border/60 bg-background px-6 py-6 lg:hidden"
			aria-label="Mobile navigation"
		>
			<div class="mx-auto max-w-[1600px] space-y-8">
				<ul class="divide-y divide-border/60">
					{#each primaryNav as n}
						<li>
							<a
								href={n.href}
								class="{mobileLinkClass} {isActive(n.href) ? 'text-foreground' : ''}"
								onclick={closeMenu}
							>
								{n.label}
							</a>
						</li>
					{/each}
				</ul>

				<ul class="divide-y divide-border/60 border-t border-border/60 pt-2">
					{#each secondaryNav as n}
						<li>
							<a
								href={n.href}
								class="{mobileLinkClass} {isActive(n.href) ? 'text-foreground' : ''}"
								onclick={closeMenu}
							>
								{n.label}
							</a>
						</li>
					{/each}
					{#if $currentUser?.role === 'artist'}
						<li>
							<a
								href="/studio"
								class="{mobileLinkClass} {isActive('/studio') ? 'text-foreground' : ''}"
								onclick={closeMenu}
							>
								Studio
							</a>
						</li>
					{/if}
					<li>
						{#if $currentUser}
							<button type="button" class="{mobileLinkClass} w-full text-left" onclick={logout}>
								Sign out ({$currentUser.email})
							</button>
						{:else}
							<a href="/login" class="{mobileLinkClass}" onclick={closeMenu}>Sign in</a>
						{/if}
					</li>
				</ul>
			</div>
		</nav>
	{/if}
</header>
