<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authService } from '$lib/application/auth';
	import { favoritesService } from '$lib/application/favorites';
	import type { Article } from '$lib/core/domain/content';

	let loading = $state(true);
	let error = $state('');
	let articles = $state<Article[]>([]);

	onMount(async () => {
		const user = await authService.load();
		if (!user) {
			await goto('/login?return_to=/wiki/saved');
			return;
		}
		try {
			articles = await favoritesService.listMine();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load saved articles';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head><title>Saved articles — Artiv Wiki</title></svelte:head>

<section class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-20">
	<a
		href="/wiki"
		class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground hover:text-foreground"
	>
		← Back to wiki
	</a>
	<p class="mt-8 font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Your library</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Saved articles</h1>

	{#if loading}
		<p class="mt-8 text-muted-foreground">Loading…</p>
	{:else if error}
		<p class="mt-8 text-sm text-destructive" role="alert">{error}</p>
	{:else if articles.length === 0}
		<p class="mt-8 text-sm text-muted-foreground">
			Nothing saved yet. Use the "Save" button on any wiki entry to keep it here.
		</p>
	{:else}
		<div class="mt-8 grid gap-6 md:grid-cols-2 lg:grid-cols-3" data-testid="web-wiki-saved-list">
			{#each articles as a}
				<a
					href="/wiki/{a.slug}"
					class="group flex flex-col justify-between rounded-sm border border-border/70 bg-card/40 p-6 transition hover:-translate-y-0.5 hover:border-foreground hover:bg-card hover:shadow-[0_12px_30px_-15px_rgba(0,0,0,0.25)]"
				>
					<div>
						<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-accent">
							{a.category ?? 'General'}
						</span>
						<h2 class="mt-4 font-display text-2xl leading-tight text-foreground">{a.title}</h2>
						<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
							{a.excerpt ?? a.body.slice(0, 140)}
						</p>
					</div>
					<div
						class="mt-6 flex items-center justify-end border-t border-border/60 pt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
					>
						<span class="transition group-hover:translate-x-1 group-hover:text-accent">Read →</span>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</section>
