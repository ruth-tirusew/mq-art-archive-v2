<script lang="ts">
	import type { ArtPost } from '$lib/core/domain/art';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import {
		acquisitionImage,
		acquisitionTitle,
		artistDiscipline,
		artistHandle,
		artistLocation,
		artistName,
		artistPortrait
	} from '$lib/utils/fields';

	type Props = {
		artist: ArtistProfile;
		works: ArtPost[];
		framed?: boolean;
		showHeader?: boolean;
		demo?: boolean;
	};

	let { artist, works, framed = false, showHeader = true, demo = false }: Props = $props();

	const handle = $derived(artistHandle(artist));

	type ContactButton = {
		label: string;
		sub: string | null;
		primary?: boolean;
		href?: string;
	};

	const contactButtons = $derived.by(() => {
		const buttons: ContactButton[] = [
			{ label: 'Request a commission', sub: null, primary: true }
		];
		if (artist.social?.telegram) {
			const href = artist.social.telegram.startsWith('http')
				? artist.social.telegram
				: `https://t.me/${artist.social.telegram.replace('@', '')}`;
			buttons.push({
				label: 'Telegram',
				sub: artist.social.telegram.replace('https://t.me/', '@'),
				href
			});
		}
		if (artist.social?.instagram) {
			const href = artist.social.instagram.startsWith('http')
				? artist.social.instagram
				: `https://instagram.com/${artist.social.instagram.replace('@', '')}`;
			buttons.push({
				label: 'Instagram',
				sub: artist.social.instagram.replace('https://instagram.com/', '@'),
				href
			});
		}
		return buttons;
	});

	const bioPreview = $derived((artist.bio ?? artist.tagline ?? '').slice(0, 110));
</script>

{#snippet profileInner()}
	<div class="rounded-[1.6rem] bg-background p-5">
		{#if showHeader}
			<div class="flex items-center justify-between font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
				<span class="truncate">makdas.et / @{handle}</span>
				<span>⌘ share</span>
			</div>
		{/if}

		<div class="mt-5 flex items-center gap-3">
			{#if artistPortrait(artist)}
				<img
					src={artistPortrait(artist)}
					alt={artistName(artist)}
					class="h-14 w-14 shrink-0 rounded-full object-cover"
				/>
			{/if}
			<div class="min-w-0">
				<div class="flex items-center gap-2">
					<p class="truncate font-display text-lg text-foreground">{artistName(artist)}</p>
				</div>
				<p class="truncate font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
					{artistDiscipline(artist) ?? ''}{#if artistLocation(artist)} · {artistLocation(artist)}{/if}
				</p>
			</div>
		</div>

		{#if bioPreview}
			<p class="mt-4 text-sm leading-relaxed text-foreground/85">{bioPreview}…</p>
		{/if}

		{#if contactButtons.length > 0}
			<div class="mt-5 space-y-2">
				{#each contactButtons as b}
					{#if b.href}
						<a
							href={b.href}
							target="_blank"
							rel="noopener noreferrer"
							class="block w-full rounded-sm border px-3 py-2.5 text-left transition {b.primary
								? 'border-accent bg-accent text-accent-foreground hover:opacity-90'
								: 'border-border bg-card/50 text-foreground hover:border-foreground'}"
						>
							<p class="font-mono text-[10px] uppercase tracking-[0.15em]">{b.label}</p>
							{#if b.sub}
								<p class="mt-0.5 font-mono text-[9px] uppercase tracking-[0.2em] opacity-70">
									{b.sub}
								</p>
							{/if}
						</a>
					{:else}
						<div
							class="block w-full rounded-sm border px-3 py-2.5 text-left {b.primary
								? 'border-accent bg-accent text-accent-foreground'
								: 'border-border bg-card/50 text-foreground'}"
						>
							<p class="font-mono text-[10px] uppercase tracking-[0.15em]">{b.label}</p>
							{#if b.sub}
								<p class="mt-0.5 font-mono text-[9px] uppercase tracking-[0.2em] opacity-70">
									{b.sub}
								</p>
							{/if}
						</div>
					{/if}
				{/each}
			</div>
		{/if}

		{#if works.length > 0}
			<div class="mt-5">
				<div class="grid grid-cols-3 gap-1.5">
					{#each works.slice(0, 3) as work}
						{@const image = acquisitionImage(work)}
						<div class="aspect-square overflow-hidden rounded-sm bg-muted">
							{#if image}
								<img src={image} alt={acquisitionTitle(work)} class="h-full w-full object-cover" />
							{/if}
						</div>
					{/each}
				</div>
				{#if demo}
					<a
						href="/portfolio"
						class="mt-3 block text-center font-mono text-[10px] uppercase tracking-[0.25em] text-accent hover:underline"
					>
						Claim your own @handle →
					</a>
				{:else}
					<a
						href="/artists/{artist.slug}"
						class="mt-3 block text-center font-mono text-[10px] uppercase tracking-[0.25em] text-accent hover:underline"
					>
						View full archive →
					</a>
				{/if}
			</div>
		{/if}
	</div>
{/snippet}

{#if framed}
	<div
		class="mx-auto max-w-[380px] overflow-hidden rounded-[2rem] border border-border/70 bg-card/40 p-2 shadow-[0_30px_80px_-30px_rgba(0,0,0,0.25)]"
	>
		{@render profileInner()}
	</div>
{:else}
	<div
		class="mx-auto max-w-md overflow-hidden rounded-[2rem] border border-border/70 bg-card/40 p-2 shadow-[0_30px_80px_-30px_rgba(0,0,0,0.25)]"
	>
		{@render profileInner()}
	</div>
{/if}
