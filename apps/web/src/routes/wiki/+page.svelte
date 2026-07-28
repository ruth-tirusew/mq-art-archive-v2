<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import VerifiedChip from '$lib/components/VerifiedChip.svelte';
	import EmptySectionPrompt from '$lib/components/home/EmptySectionPrompt.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	function filterCategory(value: string) {
		const params = new URLSearchParams($page.url.searchParams);
		if (!value || value === 'All') params.delete('category');
		else params.set('category', value);
		goto(`/wiki?${params.toString()}`);
	}

	function articleExcerpt(article: Article): string {
		return article.excerpt ?? article.body.slice(0, 140);
	}

	function articleDifficulty(article: Article): string | undefined {
		return article.difficulty;
	}

	function articleReadingTime(article: Article): number | undefined {
		return article.reading_time;
	}

	function articleContributors(article: Article): number {
		return article.contributors ?? 0;
	}

	function articleUpdated(article: Article): string | undefined {
		return article.updated_at || undefined;
	}

	const contributorTotal = $derived(
		data.articles.reduce((sum, a) => sum + articleContributors(a), 0)
	);
</script>

<svelte:head>
	<title>Wiki — Artiv</title>
	<meta
		name="description"
		content="A crowdsourced handbook for navigating life as an Ethiopian creative — legal, material, financial, and practical."
	/>
</svelte:head>

<section class="border-b border-border/60">
	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
		<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
			⁂ &nbsp; Pillar 01 · Collective knowledge
		</p>
		<h1 class="mt-4 max-w-3xl font-display text-4xl leading-[1.05] text-foreground md:text-6xl">
			The handbook for working as an artist in <em class="italic">Ethiopia</em>.
		</h1>
		<p class="mt-6 max-w-2xl text-base leading-relaxed text-muted-foreground md:text-lg">
			A crowdsourced wiki — written by the people who actually file the paperwork, walk to Mercato
			for the linen, and negotiate the contracts. Verified by a standing circle of moderators.
		</p>

		<div class="mt-10 flex flex-wrap items-center gap-3 font-mono text-[10px] uppercase tracking-[0.2em]">
			<button
				type="button"
				onclick={() => filterCategory('')}
				class="rounded-full border px-3 py-1.5 transition {!data.filterCategory
					? 'border-foreground bg-foreground text-background'
					: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
			>
				All
			</button>
			{#each data.categories as c}
				<button
					type="button"
					onclick={() => filterCategory(c)}
					class="rounded-full border px-3 py-1.5 transition {data.filterCategory === c
						? 'border-foreground bg-foreground text-background'
						: 'border-border text-muted-foreground hover:border-foreground hover:text-foreground'}"
				>
					{c}
				</button>
			{/each}
			<span class="ml-auto text-muted-foreground">
				{data.articles.length} entries · {contributorTotal} contributors
			</span>
		</div>
	</div>
</section>

<section class="mx-auto max-w-[1600px] px-6 py-14 md:px-10 md:py-16">
	{#if data.error}
		<div class="rounded-sm border border-border bg-card/40 px-6 py-12 text-center">
			<h2 class="font-display text-2xl text-foreground">Wiki temporarily unavailable</h2>
			<p class="mt-3 text-sm text-muted-foreground">We couldn't load the handbook from the API.</p>
			<button
				type="button"
				onclick={() => location.reload()}
				class="mt-6 rounded-full bg-foreground px-5 py-2.5 font-mono text-[10px] uppercase tracking-[0.2em] text-background"
			>
				Retry
			</button>
		</div>
	{:else if data.articles.length > 0}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each data.articles as a}
				{@const updated = articleUpdated(a)}
				{@const updatedLabel = updated
					? new Date(updated).toLocaleDateString('en-GB', { month: 'short', year: 'numeric' })
					: ''}
				{@const diffTone =
					articleDifficulty(a) === 'Beginner'
						? 'text-emerald-700'
						: articleDifficulty(a) === 'Intermediate'
							? 'text-ochre'
							: 'text-accent'}
				<a
					href="/wiki/{a.slug}"
					class="group flex flex-col justify-between rounded-sm border border-border/70 bg-card/40 p-6 transition hover:-translate-y-0.5 hover:border-foreground hover:bg-card hover:shadow-[0_12px_30px_-15px_rgba(0,0,0,0.25)]"
				>
					<div>
						<div class="flex items-center justify-between">
							<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
								{a.category ?? 'General'}
							</span>
							{#if a.verified}
								<VerifiedChip />
							{/if}
						</div>
						<h2 class="mt-4 font-display text-2xl leading-tight text-foreground">{a.title}</h2>
						<p class="mt-3 text-sm leading-relaxed text-muted-foreground">{articleExcerpt(a)}</p>

						<div
							class="mt-5 flex flex-wrap items-center gap-x-4 gap-y-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
						>
							{#if articleReadingTime(a)}
								<span class="inline-flex items-center gap-1.5">
									<span aria-hidden="true">◷</span> {articleReadingTime(a)} min
								</span>
							{/if}
							{#if articleDifficulty(a)}
								<span class="inline-flex items-center gap-1.5 {diffTone}">
									<span aria-hidden="true">▲</span> {articleDifficulty(a)}
								</span>
							{/if}
							{#if updatedLabel}
								<span class="inline-flex items-center gap-1.5">
									<span aria-hidden="true">↻</span> {updatedLabel}
								</span>
							{/if}
						</div>
					</div>
					<div
						class="mt-6 flex items-center justify-between border-t border-border/60 pt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
					>
						<span>{articleContributors(a)} contributors</span>
						<span class="transition group-hover:translate-x-1 group-hover:text-accent">Read →</span>
					</div>
				</a>
			{/each}
		</div>
	{:else}
		<EmptySectionPrompt
			eyebrow="Collective knowledge"
			title="No entries yet"
			body="The handbook is just starting. Anyone can propose entries on legal, material, and practical topics — written by artists who file the paperwork themselves."
			ctaLabel="Learn how to contribute"
			ctaHref="/wiki#contribute"
		/>
	{/if}
</section>

<section id="contribute" class="border-y border-border/60 bg-card/30">
	<div class="mx-auto max-w-[1600px] px-6 py-20 md:px-10 md:py-24">
		<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
			✕ &nbsp; Gatekeeping, without bottlenecks
		</p>
		<h2 class="mt-3 max-w-2xl font-display text-3xl text-foreground md:text-4xl">
			A three-tiered model that scales trust.
		</h2>

		<div class="mt-12 grid gap-6 md:grid-cols-3">
			{#each [
				{
					tier: 'Tier 01',
					name: 'Open Access',
					body: 'Anyone can read, propose edits, or draft new entries. Submissions enter a peer-review queue before going live.'
				},
				{
					tier: 'Tier 02',
					name: 'Verified Artist',
					body: 'Earn a blue check by linking an authentic portfolio or receiving a vouch from a recognised institution.'
				},
				{
					tier: 'Tier 03',
					name: 'Institutional Partner',
					body: 'Alle School, Goethe-Institut, Alliance Éthiopienne and others publish events, curriculum and legal resources directly.'
				}
			] as t}
				<div class="rounded-sm border border-border/70 bg-background p-7">
					<p class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
						{t.tier}
					</p>
					<h3 class="mt-3 font-display text-2xl text-foreground">{t.name}</h3>
					<p class="mt-4 text-sm leading-relaxed text-muted-foreground">{t.body}</p>
				</div>
			{/each}
		</div>
	</div>
</section>
