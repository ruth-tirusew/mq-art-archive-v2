<script lang="ts">
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
		postImageUrl,
		splitDisplayName
	} from '$lib/utils/fields';

	type Props = {
		artist: ArtistProfile;
		posts?: ArtPost[];
	};

	let { artist, posts = [] }: Props = $props();

	const featuredWork = $derived(posts[0]);
	const coverUrl = $derived(
		featuredWork
			? (postImageUrl(featuredWork.media, featuredWork.id, featuredWork.artist_slug) ??
				artistPortrait(artist))
			: artistPortrait(artist)
	);
	const nameParts = $derived(splitDisplayName(artistName(artist)));

	let cursor = $state({ x: 0, y: 0, visible: false });
	let heroEl: HTMLElement | undefined = $state();

	$effect(() => {
		const el = heroEl;
		if (!el) return;
		const onMove = (e: MouseEvent) => {
			const r = el.getBoundingClientRect();
			cursor = { x: e.clientX - r.left, y: e.clientY - r.top, visible: true };
		};
		const onLeave = () => {
			cursor = { ...cursor, visible: false };
		};
		el.addEventListener('mousemove', onMove);
		el.addEventListener('mouseleave', onLeave);
		return () => {
			el.removeEventListener('mousemove', onMove);
			el.removeEventListener('mouseleave', onLeave);
		};
	});

	const today = new Date().toLocaleDateString('en-GB', {
		day: '2-digit',
		month: 'short',
		year: 'numeric'
	});
</script>

<section bind:this={heroEl} class="relative overflow-hidden border-b border-border/60 bg-ink">
	<div
		class="relative z-20 mx-auto flex max-w-[1600px] items-center justify-between px-6 pt-6 font-mono text-[10px] uppercase tracking-[0.3em] text-cream/70 md:px-10"
	>
		<span class="flex items-center gap-3">
			<span class="text-accent">●</span> Featured artist · Vol. 04
		</span>
		<span class="hidden md:inline">{artistYearsActive(artist) ?? '—'}</span>
		<span>{today}</span>
	</div>

	<div
		class="relative mx-auto grid max-w-[1600px] grid-cols-12 gap-4 px-6 pt-6 pb-10 md:gap-6 md:px-10 md:pt-10 md:pb-16"
	>
		<a
			href="/artists/{artistSlug(artist)}"
			class="group relative col-span-12 aspect-[4/5] overflow-hidden rounded-sm bg-card md:col-span-8 md:aspect-[16/11]"
		>
			{#if coverUrl}
				<img
					src={coverUrl}
					alt={featuredWork?.title ?? artistName(artist)}
					class="h-full w-full object-cover transition duration-[1200ms] ease-out group-hover:scale-[1.04]"
				/>
			{/if}
			<div class="pointer-events-none absolute inset-0 bg-gradient-to-t from-ink/70 via-ink/10 to-transparent"></div>

			<div class="pointer-events-none absolute inset-x-0 bottom-0 px-5 pb-5 md:px-8 md:pb-8">
				<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-cream/70">
					Now showing / {featuredWork?.title ?? artistName(artist)}
				</p>
				<h1 class="mt-2 font-display text-[14vw] leading-[0.9] tracking-tight text-cream md:text-[7.5vw]">
					{nameParts.first}
					<br />
					<span class="italic text-cream/90">{nameParts.last}</span>
				</h1>
			</div>

			{#if cursor.visible}
				<div
					class="pointer-events-none absolute z-10 -translate-x-1/2 -translate-y-1/2 rounded-full bg-accent px-4 py-2 font-mono text-[10px] uppercase tracking-[0.2em] text-accent-foreground shadow-lg"
					style:left="{cursor.x}px"
					style:top="{cursor.y}px"
				>
					Open profile →
				</div>
			{/if}
		</a>

		<aside class="col-span-12 flex flex-col justify-between gap-8 md:col-span-4">
			<div>
				<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">01 / Featured</p>
				<p class="mt-4 font-display text-2xl leading-snug text-cream md:text-3xl">
					{artistTagline(artist)}
				</p>
			</div>

			<dl
				class="grid grid-cols-2 gap-x-6 gap-y-5 border-t border-cream/15 pt-6 font-mono text-[10px] uppercase tracking-[0.2em] text-cream/70"
			>
				{#if artistBorn(artist)}
					<div>
						<dt class="text-cream/40">Born</dt>
						<dd class="mt-1 text-cream">{artistBorn(artist)}</dd>
					</div>
				{/if}
				{#if artistLocation(artist)}
					<div>
						<dt class="text-cream/40">Based</dt>
						<dd class="mt-1 text-cream">{artistLocation(artist)}</dd>
					</div>
				{/if}
				{#if artistDiscipline(artist)}
					<div>
						<dt class="text-cream/40">Discipline</dt>
						<dd class="mt-1 text-cream">{artistDiscipline(artist)}</dd>
					</div>
				{/if}
				<div>
					<dt class="text-cream/40">Palette</dt>
					<dd class="mt-1 flex gap-1">
						{#each featuredWork?.palette ?? [] as c}
							<span class="h-4 w-2" style:background={c}></span>
						{/each}
					</dd>
				</div>
			</dl>

			<div class="flex flex-wrap items-center gap-4">
				<a
					href="/artists/{artistSlug(artist)}"
					class="group inline-flex items-center gap-3 rounded-full bg-cream px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-ink transition hover:bg-accent hover:text-accent-foreground"
				>
					Enter profile
					<span class="transition group-hover:translate-x-1">→</span>
				</a>
				<a
					href="/archive"
					class="font-mono text-[11px] uppercase tracking-[0.2em] text-cream/80 underline decoration-accent decoration-2 underline-offset-8 hover:text-cream"
				>
					All works
				</a>
			</div>
		</aside>
	</div>
</section>
