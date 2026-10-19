<script lang="ts">
	import { onMount } from 'svelte';
	import CtaLink from '$lib/components/CtaLink.svelte';
	import type { ArtPost } from '$lib/core/domain/art';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import {
		artistBorn,
		artistDiscipline,
		artistLocation,
		artistName,
		artistPortrait,
		artistSlug,
		artistTagline,
		artistYearsActive,
		postImageUrl
	} from '$lib/utils/fields';

	type Props = {
		artist: ArtistProfile;
		posts?: ArtPost[];
	};

	let { artist, posts = [] }: Props = $props();

	const works = $derived(
		posts
			.map((post) => ({ post, src: postImageUrl(post.media, post.id, post.artist_slug) }))
			.filter((w): w is { post: ArtPost; src: string } => Boolean(w.src))
			.slice(0, 5)
	);
	const portrait = $derived(artistPortrait(artist));
	const href = $derived(`/artists/${artistSlug(artist)}`);

	let current = $state(0);
	onMount(() => {
		if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
		const timer = setInterval(() => {
			if (works.length > 1) current = (current + 1) % works.length;
		}, 4500);
		return () => clearInterval(timer);
	});

	const facts = $derived(
		[
			{ label: 'Based', value: artistLocation(artist) },
			{ label: 'Discipline', value: artistDiscipline(artist) },
			{ label: 'Born', value: artistBorn(artist) },
			{ label: 'Active', value: artistYearsActive(artist) }
		].filter((f): f is { label: string; value: string } => Boolean(f.value))
	);
</script>

<section class="bg-ink text-cream" aria-labelledby="featured-artist-name">
	<div class="mx-auto grid max-w-[1600px] gap-10 px-6 py-20 md:px-10 md:py-28 lg:grid-cols-12 lg:gap-16">
		<!-- Media -->
		<a {href} class="group relative block overflow-hidden rounded-sm bg-cream/5 lg:col-span-7" tabindex="-1" aria-hidden="true">
			<div class="aspect-[4/3]">
				{#if works.length > 0}
					{#each works as work, i (work.post.id)}
						<img
							src={work.src}
							alt=""
							loading={i === 0 ? 'eager' : 'lazy'}
							class="absolute inset-0 h-full w-full object-cover transition-[opacity,transform] duration-1000 ease-out group-hover:scale-[1.03]"
							style:opacity={i === current ? 1 : 0}
						/>
					{/each}
				{:else if portrait}
					<img src={portrait} alt="" class="h-full w-full object-cover" />
				{/if}
			</div>
		</a>

		<!-- Copy -->
		<div class="flex flex-col justify-center gap-8 lg:col-span-5">
			<div class="flex flex-col gap-4">
				<p class="t-eyebrow text-ochre">Featured artist · Vol. 04</p>
				<h2 id="featured-artist-name" class="t-h2">
					<a {href} class="transition-colors hover:text-ochre">{artistName(artist)}</a>
				</h2>
				{#if artistTagline(artist)}
					<p class="t-lead text-cream/75">{artistTagline(artist)}</p>
				{/if}
			</div>

			{#if facts.length > 0}
				<dl class="grid grid-cols-2 gap-x-6 gap-y-5 border-t border-cream/15 pt-6">
					{#each facts as fact}
						<div class="flex flex-col gap-1">
							<dt class="t-eyebrow text-cream/60">{fact.label}</dt>
							<dd class="text-[15px] text-cream">{fact.value}</dd>
						</div>
					{/each}
				</dl>
			{/if}

			{#if works[current]}
				<p class="text-sm text-cream/70">
					Now showing <span class="font-display text-base text-cream italic">{works[current].post.title}</span>
					{#if works.length > 1}<span class="text-cream/60"> · {current + 1} / {works.length}</span>{/if}
				</p>
			{/if}

			<div class="flex flex-wrap items-center gap-6">
				<CtaLink {href} variant="primary" tone="on-dark">Enter profile <span aria-hidden="true">→</span></CtaLink>
				<CtaLink href="/archive" tone="on-dark">All works</CtaLink>
			</div>
		</div>
	</div>
</section>
