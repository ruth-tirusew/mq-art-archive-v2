<script lang="ts">
	import { page } from '$app/stores';
	import SearchTrigger from './GlobalSearch.svelte';

	const exploreNav = [
		{ href: '/artists', label: 'Artists' },
		{ href: '/archive', label: 'Archive' },
		{ href: '/events', label: 'Events' }
	];

	const learnNav = [{ href: '/wiki', label: 'Wiki' }];

	const secondaryNav = [
		{ href: '/about', label: 'About' },
		{ href: '/portfolio', label: 'Portfolio' }
	];

	const navLinkClass =
		'relative text-foreground/85 transition after:absolute after:-bottom-1 after:left-0 after:h-px after:w-full after:origin-left after:scale-x-0 after:bg-accent after:transition-transform after:duration-300 hover:text-foreground hover:after:scale-x-100';

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
</script>

<svelte:window onkeydown={handleKeydown} />

<header class="sticky top-0 z-40 border-b border-border/60 bg-background/85 backdrop-blur">
	<div class="mx-auto flex max-w-[1600px] items-center justify-between gap-3 px-6 py-4 md:gap-4 md:px-10">
		<a href="/" class="group flex min-w-0 items-baseline gap-2" onclick={closeMenu}>
			<span
				class="font-display text-2xl font-medium tracking-tight text-foreground transition group-hover:text-accent"
			>
				mäkdäs
			</span>
			<span class="hidden font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground sm:inline">
				/ archive · wiki · events
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

			<nav class="hidden items-center gap-x-4 gap-y-2 font-mono text-[11px] uppercase tracking-[0.2em] lg:flex">
				<div class="flex items-center gap-3">
					<span class="text-[9px] tracking-[0.3em] text-muted-foreground/70">Explore</span>
					{#each exploreNav as n}
						<a
							href={n.href}
							class="{navLinkClass} {isActive(n.href) ? 'text-foreground after:scale-x-100' : ''}"
						>
							{n.label}
						</a>
					{/each}
				</div>

				<div class="flex items-center gap-3">
					<span class="text-[9px] tracking-[0.3em] text-muted-foreground/70">Learn</span>
					{#each learnNav as n}
						<a
							href={n.href}
							class="{navLinkClass} {isActive(n.href) ? 'text-foreground after:scale-x-100' : ''}"
						>
							{n.label}
						</a>
					{/each}
				</div>

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
				<div>
					<p class="mb-3 font-mono text-[9px] uppercase tracking-[0.3em] text-muted-foreground/70">
						Explore
					</p>
					<ul class="divide-y divide-border/60">
						{#each exploreNav as n}
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
				</div>

				<div>
					<p class="mb-3 font-mono text-[9px] uppercase tracking-[0.3em] text-muted-foreground/70">
						Learn
					</p>
					<ul class="divide-y divide-border/60">
						{#each learnNav as n}
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
				</div>

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
				</ul>
			</div>
		</nav>
	{/if}
</header>
