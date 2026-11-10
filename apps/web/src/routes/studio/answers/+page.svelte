<script lang="ts">
	import { onMount } from 'svelte';
	import { questionnaireService } from '$lib/composition';
	import Skeleton from '$lib/components/Skeleton.svelte';
	import QuestionnaireForm from '$lib/components/questionnaire/QuestionnaireForm.svelte';
	import type { QuestionnaireForm as QuestionnaireFormData } from '$lib/core/domain/questionnaire';
	import { Alert, EmptyState, ErrorState, PageHeader } from '$ui';

	let form = $state<QuestionnaireFormData | null>(null);
	let highlight = $state<string[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			const [f, nudge] = await Promise.all([questionnaireService.getForm(), questionnaireService.getNudge()]);
			form = f;
			// Questions a nudge was about are flagged as new, but only if they're still unanswered.
			const answered = new Set((f?.answers ?? []).filter((a) => a.body.trim()).map((a) => a.key));
			highlight = (nudge?.questions ?? []).map((q) => q.key).filter((k) => !answered.has(k));
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load your questions';
		} finally {
			loading = false;
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Your questions — Artiv studio</title>
</svelte:head>

<section class="mx-auto px-6 py-16 md:px-10 md:py-20">
	<PageHeader
		eyebrow="Studio"
		lead="Your answers help people get to know you. They appear on your public profile, except any marked as only for our team. Changes save automatically and go live right away."
	>
		Your <em class="italic">questions</em>
	</PageHeader>

	{#if loading}
		<div class="mt-10 space-y-6" role="status" aria-live="polite">
			<span class="sr-only">Loading your questions</span>
			<Skeleton class="h-6 w-2/3" />
			<Skeleton class="h-32 w-full" />
			<Skeleton class="h-6 w-1/2" />
			<Skeleton class="h-12 w-full" />
		</div>
	{:else if error}
		<div class="mt-10"><ErrorState message={error} onretry={load} /></div>
	{:else if !form}
		<EmptyState class="mt-10" title="There are no questions yet." body="When the team adds some, you'll see them here." />
	{:else}
		{#if highlight.length > 0}
			<Alert tone="info" class="mt-8">
				{highlight.length === 1 ? 'There is 1 new question' : `There are ${highlight.length} new questions`}, marked NEW below. They're optional, so answer when you like.
			</Alert>
		{/if}
		<div class="mt-10">
			<QuestionnaireForm {form} mode="studio" {highlight} />
		</div>
	{/if}
</section>
