<script lang="ts">
	import SafeHtml from '$lib/components/SafeHtml.svelte';
	import type { PublicAnswer } from '$lib/core/domain/questionnaire';
	import { answerText, linkLabel, safeHttpUrl } from '$lib/utils/answers';

	type Props = {
		answers: PublicAnswer[];
		artistName: string;
		/** Narrow layout for the shareable /@handle page. */
		compact?: boolean;
	};

	let { answers, artistName, compact = false }: Props = $props();

	const langAttr = (a: PublicAnswer) => (a.lang === 'am' ? 'am' : undefined);
</script>

{#snippet qa(answer: PublicAnswer)}
	<div class="space-y-3">
		<dt>
			<span class={compact ? 'font-display text-lg text-foreground' : 'font-display text-xl text-foreground md:text-2xl'}>
				{answer.prompt_en}
			</span>
			{#if answer.prompt_am}
				<span class="mt-1 block text-sm text-muted-foreground" lang="am">{answer.prompt_am}</span>
			{/if}
		</dt>
		<dd class={compact ? 'text-sm leading-relaxed text-foreground/90' : 'max-w-2xl text-base leading-relaxed text-foreground/90 md:text-lg'}>
			{#if answer.format === 'rich'}
				<SafeHtml html={answer.body} lang={langAttr(answer)} />
			{:else if answer.type === 'link'}
				{@const href = safeHttpUrl(answer.body)}
				{#if href}
					<a
						{href}
						rel="noopener noreferrer nofollow ugc"
						target="_blank"
						class="break-all underline decoration-accent/60 underline-offset-4 transition hover:text-accent"
					>
						{linkLabel(href)}
					</a>
				{:else}
					<span class="break-all">{answer.body}</span>
				{/if}
			{:else}
				<p class="whitespace-pre-line" lang={langAttr(answer)}>{answerText(answer)}</p>
			{/if}
		</dd>
	</div>
{/snippet}

{#if answers.length > 0}
	{#if compact}
		<section class="mt-10 border-t border-border/60 pt-8" aria-label="In their own words" data-testid="artist-answers">
			<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">In their own words</p>
			<dl class="mt-6 space-y-8">
				{#each answers as answer (answer.key)}{@render qa(answer)}{/each}
			</dl>
		</section>
	{:else}
		<section class="border-b border-border/60" aria-label="In their own words" data-testid="artist-answers">
			<div class="mx-auto grid max-w-[1600px] grid-cols-12 gap-6 px-6 py-16 md:px-10 md:py-24">
				<div class="col-span-12 md:col-span-4">
					<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">In their own words</p>
					<h2 class="mt-4 font-display text-3xl leading-tight text-foreground md:text-4xl">
						{artistName}, answering a few questions.
					</h2>
				</div>
				<dl class="col-span-12 space-y-12 md:col-span-8">
					{#each answers as answer (answer.key)}{@render qa(answer)}{/each}
				</dl>
			</div>
		</section>
	{/if}
{/if}
