<script lang="ts">
	import CtaLink from '$lib/components/CtaLink.svelte';
	import SectionHeader from '$lib/components/home/SectionHeader.svelte';
	import type { Article } from '$lib/core/domain/content';
	import type { Event } from '$lib/core/domain/events';
	import { formatEventDate } from '$lib/utils/display';

	type Props = {
		articles: Article[];
		events: Event[];
	};

	let { articles, events }: Props = $props();

	const rows = $derived(articles.slice(0, 3));
	const nextEvents = $derived(events.slice(0, 2));

	function excerpt(article: Article): string {
		return article.excerpt ?? article.body.slice(0, 110);
	}
</script>

<section class="border-t border-border/60 bg-background" aria-labelledby="wiki-heading">
	<div class="mx-auto flex max-w-[1600px] flex-col gap-12 px-6 py-20 md:px-10 md:py-28">
		<SectionHeader eyebrow="Held in common" id="wiki-heading" action={{ href: '/wiki', label: 'Browse the wiki' }}>
			The <em class="italic">wiki.</em>
		</SectionHeader>

		{#if rows.length > 0}
			<ol class="border-t border-border">
				{#each rows as article, index (article.id)}
					<li class="border-b border-border">
						<a
							href="/wiki/{article.slug}"
							class="group grid grid-cols-[2rem_1fr_auto] items-baseline gap-x-6 gap-y-2 py-6 md:grid-cols-[3rem_9rem_minmax(0,1.2fr)_minmax(0,1fr)_auto] md:gap-x-8"
						>
							<span class="t-eyebrow text-muted-foreground">{String(index + 1).padStart(2, '0')}</span>
							<span class="t-eyebrow hidden text-accent md:block">{article.category ?? 'General'}</span>
							<span class="t-h3 text-foreground transition-colors group-hover:text-accent">{article.title}</span>
							<span class="hidden text-sm leading-relaxed text-muted-foreground md:block">{excerpt(article)}</span>
							<span
								class="text-foreground transition-transform group-hover:translate-x-1 group-hover:text-accent"
								aria-hidden="true">→</span
							>
						</a>
					</li>
				{/each}
			</ol>
		{:else}
			<p class="t-lead text-muted-foreground">
				The handbook for working as an artist in Ethiopia — EIPA registration, gallery contracts,
				pigment sources and payment rails, written by the people who file the paperwork.
			</p>
		{/if}

		<div class="grid gap-6 md:grid-cols-2">
			<!-- Primary: the conversion gets the filled card. -->
			<div class="flex flex-col gap-6 rounded-sm bg-ink p-8 text-cream md:p-10">
				<div class="flex flex-col gap-4">
					<p class="t-eyebrow text-ochre">Apply</p>
					<h3 class="t-h2">Get listed on <em class="italic text-ochre">Artiv.</em></h3>
					<p class="t-lead text-cream/75">
						Artists in Ethiopia can apply for a verified profile and share their work with the world.
					</p>
				</div>
				<div class="mt-auto">
					<CtaLink href="/apply" variant="accent" tone="on-dark">Apply now <span aria-hidden="true">→</span></CtaLink>
				</div>
			</div>

			<!-- Secondary: outlined card. -->
			<div class="flex flex-col gap-6 rounded-sm border border-border p-8 md:p-10">
				<div class="flex flex-col gap-4">
					<p class="t-eyebrow text-accent">Events</p>
					<h3 class="t-h2 text-foreground">What’s opening <em class="italic">in Addis.</em></h3>
					{#if nextEvents.length > 0}
						<ul class="divide-y divide-border border-y border-border">
							{#each nextEvents as e (e.id)}
								<li>
									<a href="/events/{e.slug}" class="group flex items-baseline justify-between gap-6 py-3">
										<span class="text-[15px] text-foreground group-hover:text-accent">{e.title}</span>
										<span class="t-eyebrow shrink-0 text-muted-foreground">{formatEventDate(e.starts_at)}</span>
									</a>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="t-lead text-muted-foreground">
							Openings, talks and residencies — verified by institutional partners.
						</p>
					{/if}
				</div>
				<div class="mt-auto">
					<CtaLink href="/events">See the calendar <span aria-hidden="true">→</span></CtaLink>
				</div>
			</div>
		</div>
	</div>
</section>
