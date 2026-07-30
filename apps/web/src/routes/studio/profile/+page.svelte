<script lang="ts">
	import { onMount } from 'svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import MediaUploader from '$lib/components/MediaUploader.svelte';
	import { profileService } from '$lib/application/profiles';
	import { artPostService } from '$lib/application/artPosts';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import type { ArtPost } from '$lib/core/domain/art';
	import type { UpdateMyProfileInput } from '$lib/core/ports/profile';

	let profile = $state<ArtistProfile | null>(null);
	let previewPosts = $state<ArtPost[]>([]);
	let loaded = $state(false);
	let saving = $state(false);
	let message = $state('');
	let error = $state('');

	// Flat draft fields — source of truth for both inputs and live preview.
	let displayName = $state('');
	let slug = $state('');
	let handle = $state('');
	let bio = $state('');
	let born = $state('');
	let discipline = $state('');
	let tagline = $state('');
	let yearsActive = $state('');
	let portraitUrl = $state('');
	let contactEmail = $state('');
	let contactPhone = $state('');
	let contactWebsite = $state('');
	let contactLocation = $state('');
	let socialInstagram = $state('');
	let socialTwitter = $state('');
	let socialTelegram = $state('');
	let influencesText = $state('');
	let inResidence = $state(false);
	let residencyPlace = $state('');
	let openForCommission = $state(false);
	let visibility = $state<'draft' | 'pending'>('draft');

	function applyProfile(p: ArtistProfile) {
		displayName = p.display_name;
		slug = p.slug;
		handle = p.handle ?? p.slug;
		bio = p.bio ?? '';
		born = p.born ?? '';
		discipline = p.discipline ?? '';
		tagline = p.tagline ?? '';
		yearsActive = p.years_active ?? '';
		portraitUrl = p.portrait_url ?? '';
		influencesText = (p.influences ?? []).join(', ');
		inResidence = p.in_residence ?? false;
		residencyPlace = p.residency_place ?? '';
		openForCommission = p.open_for_commission ?? false;
		contactEmail = p.contact?.email ?? '';
		contactPhone = p.contact?.phone ?? '';
		contactWebsite = p.contact?.website ?? '';
		contactLocation = p.contact?.location ?? '';
		socialInstagram = p.social?.instagram ?? '';
		socialTwitter = p.social?.twitter ?? '';
		socialTelegram = p.social?.telegram ?? '';
		visibility = p.status === 'pending' ? 'pending' : 'draft';
	}

	function parseInfluences(text: string): string[] {
		return text
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean);
	}

	function toInput(): UpdateMyProfileInput {
		const input: UpdateMyProfileInput = {
			display_name: displayName,
			slug,
			handle,
			bio,
			born,
			discipline,
			tagline,
			years_active: yearsActive,
			portrait_url: portraitUrl,
			influences: parseInfluences(influencesText),
			in_residence: inResidence,
			residency_place: residencyPlace,
			open_for_commission: openForCommission,
			contact: {
				email: contactEmail,
				phone: contactPhone,
				website: contactWebsite,
				location: contactLocation
			},
			social: {
				instagram: socialInstagram,
				twitter: socialTwitter,
				telegram: socialTelegram
			}
		};
		// Leave approved profiles approved unless the artist is still in onboarding.
		if (profile?.status !== 'approved') {
			input.status = visibility;
		}
		return input;
	}

	onMount(async () => {
		try {
			profile = await profileService.getMyProfile();
			applyProfile(profile);
			previewPosts = await artPostService.listMyPosts().catch(() => []);
			loaded = true;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Could not load profile';
		}
	});

	const previewProfile = $derived.by((): ArtistProfile | null => {
		if (!profile || !loaded) return null;
		return {
			...profile,
			display_name: displayName || profile.display_name,
			slug: slug || profile.slug,
			handle: handle || profile.handle || profile.slug,
			bio,
			born,
			discipline,
			tagline,
			years_active: yearsActive,
			portrait_url: portraitUrl,
			influences: parseInfluences(influencesText),
			in_residence: inResidence,
			residency_place: residencyPlace,
			open_for_commission: openForCommission,
			contact: {
				email: contactEmail,
				phone: contactPhone,
				website: contactWebsite,
				location: contactLocation
			},
			social: {
				instagram: socialInstagram,
				twitter: socialTwitter,
				telegram: socialTelegram
			},
			status: profile.status === 'approved' ? 'approved' : visibility
		};
	});

	const previewWorks = $derived(previewPosts.slice(0, 6));

	async function save() {
		saving = true;
		message = '';
		error = '';
		try {
			profile = await profileService.updateMyProfile(toInput());
			applyProfile(profile);
			message = 'Profile saved.';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Save failed';
		} finally {
			saving = false;
		}
	}
</script>

<section class="mx-auto max-w-[1600px] px-6 py-14 md:px-10 md:py-20">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Studio · Profile</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Edit profile</h1>

	{#if error && !loaded}
		<p class="mt-6 text-sm text-destructive" role="alert">{error}</p>
	{:else if !loaded}
		<p class="mt-6 text-muted-foreground">Loading…</p>
	{:else}
		<div class="mt-10 grid gap-10 lg:grid-cols-2">
			<form class="space-y-6" onsubmit={(e) => { e.preventDefault(); void save(); }}>
				<div>
					<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="display_name">Display name</label>
					<input id="display_name" class="field" bind:value={displayName} required />
				</div>
				<div class="grid gap-4 sm:grid-cols-2">
					<div>
						<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="handle">@handle</label>
						<input id="handle" class="field" bind:value={handle} />
					</div>
					<div>
						<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="slug">Slug</label>
						<input id="slug" class="field" bind:value={slug} />
					</div>
				</div>
				<div>
					<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="tagline">Tagline</label>
					<input id="tagline" class="field" bind:value={tagline} />
				</div>
				<div>
					<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="bio">Bio</label>
					<textarea id="bio" class="field min-h-28" bind:value={bio}></textarea>
				</div>
				<div class="grid gap-4 sm:grid-cols-2">
					<div>
						<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="discipline">Discipline</label>
						<input id="discipline" class="field" bind:value={discipline} />
					</div>
					<div>
						<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="years_active">Years active</label>
						<input id="years_active" class="field" bind:value={yearsActive} />
					</div>
				</div>
				<div>
					<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="born">Born</label>
					<input id="born" class="field" bind:value={born} placeholder="e.g. 1989, Bahir Dar" />
				</div>
				<div>
					<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="influences">Influences</label>
					<input
						id="influences"
						class="field"
						bind:value={influencesText}
						placeholder="Comma-separated, e.g. Afewerk Tekle, Skunder Boghossian"
					/>
				</div>
				<div>
					<MediaUploader onUploaded={(media) => (portraitUrl = media.secure_url)} />
					{#if portraitUrl}
						<label class="mt-3 mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="portrait_url">Portrait URL</label>
						<input id="portrait_url" class="field" value={portraitUrl} readonly />
					{/if}
				</div>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="px-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Availability</legend>
					<label class="flex items-center gap-3 text-sm text-foreground">
						<input type="checkbox" bind:checked={inResidence} />
						In residence abroad
					</label>
					{#if inResidence}
						<input class="field" placeholder="Residency place" bind:value={residencyPlace} />
					{/if}
					<label class="flex items-center gap-3 text-sm text-foreground">
						<input type="checkbox" bind:checked={openForCommission} />
						Open for commission
					</label>
				</fieldset>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="px-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Contact</legend>
					<input class="field" placeholder="Email" bind:value={contactEmail} />
					<input class="field" placeholder="Phone" bind:value={contactPhone} />
					<input class="field" placeholder="Website" bind:value={contactWebsite} />
					<input class="field" placeholder="Location" bind:value={contactLocation} />
				</fieldset>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="px-1 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Social</legend>
					<input class="field" placeholder="Instagram" bind:value={socialInstagram} />
					<input class="field" placeholder="Twitter" bind:value={socialTwitter} />
					<input class="field" placeholder="Telegram" bind:value={socialTelegram} />
				</fieldset>

				{#if profile?.status === 'approved'}
					<p class="rounded-sm border border-border/60 px-4 py-3 text-sm text-muted-foreground">
						Your profile is published. Changes appear in the live preview immediately; save to update the public page.
					</p>
				{:else}
					<div>
						<label class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground" for="status">Visibility</label>
						<select id="status" class="field" bind:value={visibility}>
							<option value="draft">Draft (only you)</option>
							<option value="pending">Submit for review</option>
						</select>
					</div>
				{/if}

				{#if message}
					<p class="text-sm text-accent">{message}</p>
				{/if}
				{#if error}
					<p class="text-sm text-destructive" role="alert">{error}</p>
				{/if}

				<button
					type="submit"
					class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90 disabled:opacity-50"
					disabled={saving}
				>
					{saving ? 'Saving…' : 'Save profile'}
				</button>
			</form>

			{#if previewProfile}
				<div class="lg:sticky lg:top-8 lg:self-start">
					<p class="mb-4 font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground">
						Live preview
					</p>
					<ShareableProfile artist={previewProfile} works={previewWorks} framed />
				</div>
			{/if}
		</div>
	{/if}
</section>

<style>
	.field {
		width: 100%;
		border: 1px solid color-mix(in oklab, var(--border) 70%, transparent);
		background: color-mix(in oklab, var(--card) 30%, transparent);
		padding: 0.75rem 1rem;
		font-size: 0.875rem;
		color: var(--foreground);
	}
	.field:focus {
		outline: 2px solid color-mix(in oklab, var(--accent) 50%, transparent);
		outline-offset: 2px;
	}
</style>
