<script lang="ts">
	import { onMount } from 'svelte';
	import { onboardingService } from '$lib/application/onboarding';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import type { OnboardingApplication } from '$lib/core/domain/onboarding';
	import { ApiError } from '$lib/adapters/api/client';

	let application = $state<OnboardingApplication | null>(null);
	let loading = $state(true);
	let error = $state('');

	const steps = ['Submitted', 'In review', 'Decision'];
	const headline = {
		pending: 'We’re reviewing it',
		approved: 'You’re approved',
		rejected: 'Not approved this time'
	};
	const stage = $derived(application ? (application.status === 'pending' ? 1 : 2) : 0);

	onMount(async () => {
		// Sign-in and artist redirects are handled by the /apply layout.
		try {
			application = await onboardingService.getMyApplication();
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) {
				error = 'No application found yet.';
			} else {
				error = e instanceof Error ? e.message : 'Could not load application';
			}
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Application status — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-2xl px-6 py-24 md:px-10">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Onboarding</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Application status</h1>

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
	{:else if error}
		<p class="mt-6 text-sm text-muted-foreground">{error}</p>
		<a href="/apply" class="mt-6 inline-block text-accent underline underline-offset-4">Submit an application</a>
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
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Your application</p>
			<p class="mt-2 font-display text-2xl text-foreground">{headline[application.status]}</p>
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
				<a
					href="/studio"
					class="mt-4 inline-flex rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90"
				>
					Set up my studio
				</a>
			{/if}
		{:else if application.status === 'rejected'}
			<p class="mt-6 text-sm text-muted-foreground">
				Your application wasn't approved this time. You can update your details and apply again.
			</p>
			<a href="/apply" class="mt-4 inline-block text-accent underline underline-offset-4">Apply again</a>
		{:else}
			<p class="mt-6 text-sm text-muted-foreground">
				Thanks for applying. Most applications are reviewed within a few days, and we’ll email you as soon
				as there’s a decision. Meanwhile, feel free to explore the site.
			</p>
			<a href="/" class="mt-4 inline-block text-accent underline underline-offset-4">Browse Artiv</a>
		{/if}
	{/if}
</section>
