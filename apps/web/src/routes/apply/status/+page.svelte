<script lang="ts">
	import { onMount } from 'svelte';
	import { onboardingService } from '$lib/application/onboarding';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import type { OnboardingApplication } from '$lib/core/domain/onboarding';
	import { ApiError } from '$lib/adapters/api/client';
	import { Badge, Button, EmptyState, ErrorState, PageHeader } from '$ui';

	let application = $state<OnboardingApplication | null>(null);
	let loading = $state(true);
	let error = $state('');
	let notFound = $state(false);

	const steps = ['Submitted', 'In review', 'Decision'];
	const headline = {
		pending: 'We’re reviewing it',
		approved: 'You’re approved',
		rejected: 'Not approved this time'
	};
	const stage = $derived(application ? (application.status === 'pending' ? 1 : 2) : 0);

	async function load() {
		// Sign-in and artist redirects are handled by the /apply layout.
		loading = true;
		error = '';
		notFound = false;
		try {
			application = await onboardingService.getMyApplication();
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) {
				notFound = true;
			} else {
				error = e instanceof Error ? e.message : 'Could not load application';
			}
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Application status — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-2xl px-6 py-24 md:px-10">
	<PageHeader eyebrow="Onboarding">Application <em class="italic">status</em></PageHeader>

	{#if loading}
		<div class="mt-8" role="status" aria-live="polite">
			<span class="sr-only">Loading your application</span>
			<div class="grid grid-cols-3 gap-2">
				<Skeleton class="h-6 w-full" />
				<Skeleton class="h-6 w-full" />
				<Skeleton class="h-6 w-full" />
			</div>
			<Skeleton class="mt-6 h-40 w-full" />
		</div>
	{:else if notFound}
		<EmptyState class="mt-8" title="You haven’t applied yet." body="Tell us about your work and our team will take a look.">
			{#snippet action()}<Button variant="secondary" size="sm" href="/apply">Start an application →</Button>{/snippet}
		</EmptyState>
	{:else if error}
		<div class="mt-8"><ErrorState message={error} onretry={load} /></div>
	{:else if application}
		<ol class="mt-8 grid grid-cols-3 gap-2 font-mono text-[10px] uppercase tracking-[0.2em]">
			{#each steps as label, i}
				<li
					class="border-t-2 pt-3 {i <= stage
						? application.status === 'rejected' && i === 2
							? 'border-destructive text-destructive'
							: 'border-accent text-foreground'
						: 'border-border text-muted-foreground'}"
				>
					{i === 2 && application.status === 'rejected' ? 'Not approved' : label}
				</li>
			{/each}
		</ol>

		<div class="mt-6 rounded-sm border border-border/70 bg-card/30 p-6">
			<div class="flex items-center justify-between gap-3">
				<p class="t-eyebrow text-muted-foreground">Your application</p>
				<Badge tone={application.status === 'approved' ? 'success' : application.status === 'rejected' ? 'destructive' : 'warning'}>{application.status}</Badge>
			</div>
			<p class="t-h3 mt-2">{headline[application.status]}</p>
			<p class="mt-4 text-sm text-foreground">{application.display_name}</p>
			{#if application.requested_handle}
				<p class="mt-2 font-mono text-xs text-accent">@{application.requested_handle}</p>
			{/if}
			{#if application.notes}
				<p class="mt-3 text-sm text-muted-foreground whitespace-pre-wrap">{application.notes}</p>
			{/if}
		</div>

		{#if application.review_notes}
			<div class="mt-4 rounded-sm border border-border/70 p-4">
				<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Note from the team</p>
				<p class="mt-2 whitespace-pre-wrap text-sm text-foreground">{application.review_notes}</p>
			</div>
		{/if}

		{#if application.status === 'approved'}
			{#if application.applicant_type === 'institution'}
				<p class="mt-6 text-sm text-muted-foreground">
					Your institution application was approved and your profile has been created. The team will be in
					touch about next steps.
				</p>
			{:else}
				<p class="mt-6 text-sm text-muted-foreground">
					Welcome to Artiv! Next, set up your profile and publish it.
				</p>
				<div class="mt-4"><Button href="/studio">Set up my studio</Button></div>
			{/if}
		{:else if application.status === 'rejected'}
			<p class="mt-6 text-sm text-muted-foreground">
				Your application wasn't approved this time. You can update your details and apply again.
			</p>
			<div class="mt-4"><Button variant="secondary" href="/apply">Apply again</Button></div>
		{:else}
			<p class="mt-6 text-sm text-muted-foreground">
				Thanks for applying. Most applications are reviewed within a few days, and we’ll email you as soon
				as there’s a decision. Meanwhile, feel free to explore the site.
			</p>
			<div class="mt-4"><Button variant="link" href="/">Browse Artiv →</Button></div>
		{/if}
	{/if}
</section>
