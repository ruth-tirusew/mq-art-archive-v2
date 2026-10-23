<script lang="ts">
	import { onMount } from 'svelte';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { profileService } from '$lib/application/profiles';
	import { artPostService } from '$lib/application/artPosts';
	import { analyticsService } from '$lib/application/analytics';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import type { ArtPost } from '$lib/core/domain/art';
	import { Alert, Badge, Button, ErrorState, PageHeader } from '$ui';

	let profile = $state<ArtistProfile | null>(null);
	let posts = $state<ArtPost[]>([]);
	let profileViews = $state<number | null>(0);
	let postsFailed = $state(false);
	let error = $state('');
	let loaded = $state(false);

	async function load() {
		error = '';
		postsFailed = false;
		try {
			const [p, ps] = await Promise.all([
				profileService.getMyProfile(),
				artPostService.listMyPosts().catch(() => {
					postsFailed = true;
					return [] as ArtPost[];
				})
			]);
			profile = p;
			posts = ps;
			// Views are a nice-to-have: show a dash when they fail rather than a false zero.
			const views = await analyticsService.query('artist', p.id).catch(() => null);
			profileViews = views ? views.reduce((total, view) => total + view.count, 0) : null;
			loaded = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load profile';
		}
	}

	onMount(load);

	const counts = $derived.by(() => {
		const drafts = posts.filter((p) => (p.status ?? 'draft') === 'draft').length;
		const published = posts.filter((p) => p.status === 'published').length;
		const archived = posts.filter((p) => p.status === 'archived').length;
		return { drafts, published, archived };
	});

	const isPublic = $derived(profile?.status === 'approved');
	const firstRun = $derived(!isPublic && posts.length === 0 && !postsFailed);
	const publishReady = $derived(Boolean(profile?.bio) && Boolean(profile?.portrait_url));

	const checklist = $derived.by(() => {
		if (!profile) return [];
		const edit = '/studio/profile';
		const items = [
			{ key: 'portrait', label: 'Portrait', ok: Boolean(profile.portrait_url), href: edit },
			{ key: 'bio', label: 'Bio', ok: Boolean(profile.bio), href: edit },
			{ key: 'tagline', label: 'Tagline', ok: Boolean(profile.tagline), href: edit },
			{ key: 'discipline', label: 'Discipline', ok: Boolean(profile.discipline), href: edit },
			{ key: 'location', label: 'Location', ok: Boolean(profile.contact?.location), href: edit },
			{ key: 'influences', label: 'Influences', ok: (profile.influences?.length ?? 0) > 0, href: edit },
			{
				key: 'published',
				label: 'At least one published work',
				ok: counts.published > 0,
				href: '/studio/posts'
			},
			{ key: 'public', label: 'Profile is public', ok: isPublic, href: edit }
		];
		if (profile.open_for_commission) {
			items.push({
				key: 'email',
				label: 'Contact email (for commissions)',
				ok: Boolean(profile.contact?.email),
				href: edit
			});
		}
		return items;
	});

	const completeness = $derived.by(() => {
		if (checklist.length === 0) return 0;
		const done = checklist.filter((c) => c.ok).length;
		return Math.round((done / checklist.length) * 100);
	});
</script>

<section class="mx-auto max-w-3xl px-6 py-14 md:px-10 md:py-20">
	<PageHeader eyebrow="Studio">
		{#if firstRun && profile}
			Welcome to <em class="italic">Artiv</em>, {profile.display_name}
		{:else}
			Welcome <em class="italic">back</em>
		{/if}
	</PageHeader>

	{#if error}
		<div class="mt-8"><ErrorState message={error} onretry={load} /></div>
	{:else if !loaded || !profile}
		<div class="mt-8 space-y-4" role="status" aria-live="polite">
			<span class="sr-only">Loading your profile</span>
			<Skeleton class="h-6 w-40" />
			<Skeleton class="h-20 w-full" />
			<Skeleton class="h-48 w-full" />
		</div>
	{:else}
		{#if postsFailed}
			<Alert tone="warning" title="Couldn’t load your works" class="mt-8">
				<div class="flex flex-wrap items-center gap-4">
					<span>Counts below may be incomplete.</span>
					<Button variant="secondary" size="sm" onclick={load}>Try again</Button>
				</div>
			</Alert>
		{/if}

		{#if !isPublic}
			<div class="mt-8 rounded-sm border border-accent/40 bg-accent/5 p-6">
				<p class="font-display text-xl text-foreground">
					{firstRun ? "You're approved — let's get you set up" : 'Your profile is still private'}
				</p>
				<p class="mt-2 text-sm leading-relaxed text-muted-foreground">
					Your profile is only visible to you for now. Three quick steps and you're live:
				</p>
				<ol class="mt-4 space-y-3 text-sm">
					<li class="flex items-start gap-3">
						<span class={profile.portrait_url && profile.bio ? 'text-accent' : 'text-muted-foreground'}>
							{profile.portrait_url && profile.bio ? '●' : '1'}
						</span>
						<span>
							<a href="/studio/profile" class="text-foreground underline underline-offset-4">
								Add a portrait and a short bio
							</a>
							<span class="block text-xs text-muted-foreground">Required before you can go public.</span>
						</span>
					</li>
					<li class="flex items-start gap-3">
						<span class={counts.published > 0 ? 'text-accent' : 'text-muted-foreground'}>
							{counts.published > 0 ? '●' : '2'}
						</span>
						<span>
							<a href="/studio/posts" class="text-foreground underline underline-offset-4">
								Add and publish your first work
							</a>
							<span class="block text-xs text-muted-foreground">Optional, but profiles with work get noticed.</span>
						</span>
					</li>
					<li class="flex items-start gap-3">
						<span class="text-muted-foreground">3</span>
						<span>
							<a href="/studio/profile" class="text-foreground underline underline-offset-4">
								Make your profile public
							</a>
							<span class="block text-xs text-muted-foreground">
								{publishReady
									? "You're ready — tick “Make my profile public” and save."
									: 'Available once step 1 is done. No further review needed.'}
							</span>
						</span>
					</li>
				</ol>
			</div>
		{/if}

		<div class="mt-8 flex flex-wrap gap-2">
			<Badge tone={isPublic ? 'success' : 'neutral'}>{isPublic ? 'Public' : (profile.status ?? 'draft')}</Badge>
			{#if profile.featured}<Badge tone="accent">Featured</Badge>{/if}
			{#if profile.open_for_commission}<Badge>Open for commission</Badge>{/if}
			{#if profile.in_residence}
				<Badge>In residence{#if profile.residency_place} · {profile.residency_place}{/if}</Badge>
			{/if}
		</div>

		{#if !firstRun}
		<div class="mt-8 flex items-center justify-between rounded-sm border border-border/70 bg-card/30 px-6 py-4">
			<div>
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Profile visits</p>
				<p class="mt-1 text-xs text-muted-foreground">Unique daily views in the last month</p>
			</div>
			<p class="font-display text-3xl text-foreground" title={profileViews === null ? 'Visit counts are unavailable right now' : undefined}>{profileViews === null ? '—' : profileViews.toLocaleString()}</p>
		</div>
		{/if}

		<div class="mt-8 rounded-sm border border-border/70 bg-card/30 p-6">
			<div class="flex items-baseline justify-between gap-4">
				<div>
					<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
						Profile completeness
					</p>
					<p class="mt-2 font-display text-3xl text-foreground">{completeness}%</p>
				</div>
				<p class="text-sm text-muted-foreground">{profile.display_name}</p>
			</div>
			<ul class="mt-6 space-y-2">
				{#each checklist as item}
					<li class="flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.15em]">
						<span class={item.ok ? 'text-accent' : 'text-muted-foreground'}>{item.ok ? '●' : '○'}</span>
						{#if item.ok}
							<span class="text-foreground">{item.label}</span>
						{:else}
							<a href={item.href} class="text-muted-foreground underline-offset-4 hover:text-foreground hover:underline">
								{item.label}
							</a>
						{/if}
					</li>
				{/each}
			</ul>
		</div>

		{#if !firstRun && !postsFailed}
		<div class="mt-6 grid grid-cols-3 gap-3">
			<div class="rounded-sm border border-border/60 p-4 text-center">
				<p class="font-display text-2xl text-foreground">{counts.drafts}</p>
				<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">Drafts</p>
			</div>
			<div class="rounded-sm border border-border/60 p-4 text-center">
				<p class="font-display text-2xl text-foreground">{counts.published}</p>
				<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">Published</p>
			</div>
			<div class="rounded-sm border border-border/60 p-4 text-center">
				<p class="font-display text-2xl text-foreground">{counts.archived}</p>
				<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">Archived</p>
			</div>
		</div>
		{/if}

		<div class="mt-8 grid gap-4 sm:grid-cols-2">
			<a
				href="/studio/profile"
				class="rounded-sm border border-border/70 p-5 transition hover:border-foreground/20 hover:bg-card/40"
			>
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">Edit</p>
				<p class="mt-2 font-display text-xl text-foreground">Profile</p>
				<p class="mt-2 text-sm text-muted-foreground">Bio, contact links, portrait, and @handle.</p>
			</a>
			<a
				href="/studio/posts"
				class="rounded-sm border border-border/70 p-5 transition hover:border-foreground/20 hover:bg-card/40"
			>
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">Manage</p>
				<p class="mt-2 font-display text-xl text-foreground">Posts</p>
				<p class="mt-2 text-sm text-muted-foreground">Draft, publish, and archive your works.</p>
			</a>
		</div>

		{#if profile.handle || profile.slug}
			<div class="mt-10">
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
					Public profile
				</p>
				{#if isPublic}
					<a
						href="/@{profile.handle || profile.slug}"
						class="mt-2 inline-block text-accent underline underline-offset-4"
					>
						View @{profile.handle || profile.slug}
					</a>
				{:else}
					<p class="mt-2 text-sm text-muted-foreground">
						@{profile.handle || profile.slug} — not public yet.
					</p>
				{/if}
			</div>
		{/if}
	{/if}
</section>
