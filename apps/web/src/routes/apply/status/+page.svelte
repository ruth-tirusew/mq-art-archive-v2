<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { onboardingService } from '$lib/application/onboarding';
	import { authService } from '$lib/application/auth';
	import type { OnboardingApplication } from '$lib/core/domain/onboarding';
	import { ApiError } from '$lib/adapters/api/client';

	let application = $state<OnboardingApplication | null>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			await goto('/login?return_to=/apply/status');
			return;
		}
		if (user.role === 'artist') {
			await goto('/studio');
			return;
		}

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
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else if error}
		<p class="mt-6 text-sm text-muted-foreground">{error}</p>
		<a href="/apply" class="mt-6 inline-block text-accent underline underline-offset-4">Submit an application</a>
	{:else if application}
		<div class="mt-8 rounded-sm border border-border/70 bg-card/30 p-6">
			<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">Status</p>
			<p class="mt-2 font-display text-2xl capitalize text-foreground">{application.status}</p>
			<p class="mt-4 text-sm text-foreground">{application.display_name}</p>
			{#if application.requested_handle}
				<p class="mt-2 font-mono text-xs text-accent">@{application.requested_handle}</p>
			{/if}
			{#if application.notes}
				<p class="mt-3 text-sm text-muted-foreground whitespace-pre-wrap">{application.notes}</p>
			{/if}
		</div>

		{#if application.status === 'approved'}
			<p class="mt-6 text-sm text-muted-foreground">
				Your application was approved. Sign out and back in if studio access does not appear immediately.
			</p>
			<a href="/studio" class="mt-4 inline-block text-accent underline underline-offset-4">Go to studio</a>
		{:else if application.status === 'rejected'}
			<p class="mt-6 text-sm text-muted-foreground">Contact the team if you believe this was a mistake.</p>
		{:else}
			<p class="mt-6 text-sm text-muted-foreground">Your application is in the review queue.</p>
		{/if}
	{/if}
</section>
