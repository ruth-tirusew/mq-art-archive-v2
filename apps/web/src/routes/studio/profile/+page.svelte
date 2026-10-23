<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { onMount } from 'svelte';
	import ShareableProfile from '$lib/components/ShareableProfile.svelte';
	import MediaUploader from '$lib/components/MediaUploader.svelte';
	import { profileService } from '$lib/application/profiles';
	import { artPostService } from '$lib/application/artPosts';
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import type { ArtPost } from '$lib/core/domain/art';
	import { Alert, Button, Checkbox, ConfirmDialog, createConfirm, ErrorState, Field, Input, PageHeader, Textarea, toast } from '$ui';
	import type { UpdateMyProfileInput } from '$lib/core/ports/profile';

	let profile = $state<ArtistProfile | null>(null);
	let previewPosts = $state<ArtPost[]>([]);
	let loaded = $state(false);
	let saving = $state(false);
	let error = $state('');
	let loadError = $state('');
	let fieldErrors = $state<{ displayName?: string; handle?: string; slug?: string }>({});
	const confirm = createConfirm();

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
	let visibility = $state<'draft' | 'approved'>('draft');

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
		visibility = p.status === 'approved' ? 'approved' : 'draft';
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
		input.status = visibility;
		return input;
	}

	async function load() {
		loadError = '';
		try {
			profile = await profileService.getMyProfile();
			applyProfile(profile);
			previewPosts = await artPostService.listMyPosts().catch(() => []);
			loaded = true;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Could not load profile';
		}
	}

	onMount(load);

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
			status: visibility
		};
	});

	const previewWorks = $derived(previewPosts.slice(0, 6));

	const noSpaces = (v: string) => (/\s/.test(v) ? 'No spaces. Use letters, numbers, dashes or underscores.' : undefined);

	function requestSave() {
		error = '';
		fieldErrors = {
			displayName: displayName.trim() ? undefined : 'Your name is required.',
			handle: noSpaces(handle),
			slug: noSpaces(slug)
		};
		if (Object.values(fieldErrors).some(Boolean)) return;
		if (profile?.status === 'approved' && visibility === 'draft') {
			confirm.ask({
				title: 'Take your profile offline?',
				description: 'Your public page and works are hidden until you make the profile public again.',
				confirmLabel: 'Save and unpublish',
				destructive: true,
				run: save
			});
			return;
		}
		void save();
	}

	async function save() {
		saving = true;
		error = '';
		try {
			profile = await profileService.updateMyProfile(toInput());
			applyProfile(profile);
			toast.success(visibility === 'approved' ? 'Profile saved and public.' : 'Profile saved.');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Save failed';
		} finally {
			saving = false;
		}
	}
</script>

<section class="container-page py-14 md:py-20">
	<PageHeader eyebrow="Studio · Profile">Edit <em class="italic">profile</em></PageHeader>

	{#if loadError}
		<div class="mt-8 max-w-2xl"><ErrorState message={loadError} onretry={load} /></div>
	{:else if !loaded}
		<div class="mt-8 space-y-4" role="status" aria-live="polite">
			<span class="sr-only">Loading profile</span>
			<Skeleton class="h-12 w-full" />
			<Skeleton class="h-12 w-full" />
			<Skeleton class="h-40 w-full" />
		</div>
	{:else}
		<div class="mt-10 grid gap-10 lg:grid-cols-2">
			<form class="space-y-6" onsubmit={(e) => { e.preventDefault(); requestSave(); }} novalidate>
				<Field label="Display name" required error={fieldErrors.displayName}>
					{#snippet children({ id, describedby, invalid })}
						<Input {id} aria-describedby={describedby} {invalid} bind:value={displayName} />
					{/snippet}
				</Field>
				<div class="grid gap-4 sm:grid-cols-2">
					<Field label="@handle" error={fieldErrors.handle}>
						{#snippet children({ id, describedby, invalid })}
							<Input {id} aria-describedby={describedby} {invalid} bind:value={handle} />
						{/snippet}
					</Field>
					<Field label="Slug" hint="Part of your artist page address." error={fieldErrors.slug}>
						{#snippet children({ id, describedby, invalid })}
							<Input {id} aria-describedby={describedby} {invalid} bind:value={slug} />
						{/snippet}
					</Field>
				</div>
				<Field label="Tagline">
					{#snippet children({ id })}<Input {id} bind:value={tagline} />{/snippet}
				</Field>
				<Field label="Bio">
					{#snippet children({ id })}<Textarea {id} rows={5} bind:value={bio} />{/snippet}
				</Field>
				<div class="grid gap-4 sm:grid-cols-2">
					<Field label="Discipline">
						{#snippet children({ id })}<Input {id} bind:value={discipline} />{/snippet}
					</Field>
					<Field label="Years active">
						{#snippet children({ id })}<Input {id} bind:value={yearsActive} />{/snippet}
					</Field>
				</div>
				<Field label="Born">
					{#snippet children({ id })}<Input {id} bind:value={born} placeholder="e.g. 1989, Bahir Dar" />{/snippet}
				</Field>
				<Field label="Influences" hint="Comma-separated, e.g. Afewerk Tekle, Skunder Boghossian">
					{#snippet children({ id, describedby })}<Input {id} aria-describedby={describedby} bind:value={influencesText} />{/snippet}
				</Field>
				<div class="space-y-3">
					<p class="t-eyebrow text-foreground">Portrait</p>
					<MediaUploader onUploaded={(media) => (portraitUrl = media.secure_url)} />
					{#if portraitUrl}
						<img src={portraitUrl} alt="Current portrait" class="h-32 w-32 rounded-sm border border-border object-cover" />
					{/if}
				</div>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="t-eyebrow px-1 text-muted-foreground">Availability</legend>
					<Checkbox bind:checked={inResidence}>In residence abroad</Checkbox>
					{#if inResidence}
						<Field label="Residency place">
							{#snippet children({ id })}<Input {id} bind:value={residencyPlace} />{/snippet}
						</Field>
					{/if}
					<Checkbox bind:checked={openForCommission}>Open for commission</Checkbox>
				</fieldset>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="t-eyebrow px-1 text-muted-foreground">Contact</legend>
					<Field label="Email">{#snippet children({ id })}<Input {id} type="email" bind:value={contactEmail} />{/snippet}</Field>
					<Field label="Phone">{#snippet children({ id })}<Input {id} type="tel" bind:value={contactPhone} />{/snippet}</Field>
					<Field label="Website">{#snippet children({ id })}<Input {id} type="url" placeholder="https://" bind:value={contactWebsite} />{/snippet}</Field>
					<Field label="Location">{#snippet children({ id })}<Input {id} bind:value={contactLocation} />{/snippet}</Field>
				</fieldset>

				<fieldset class="space-y-4 rounded-sm border border-border/60 p-4">
					<legend class="t-eyebrow px-1 text-muted-foreground">Social</legend>
					<Field label="Instagram">{#snippet children({ id })}<Input {id} bind:value={socialInstagram} />{/snippet}</Field>
					<Field label="Twitter">{#snippet children({ id })}<Input {id} bind:value={socialTwitter} />{/snippet}</Field>
					<Field label="Telegram">{#snippet children({ id })}<Input {id} bind:value={socialTelegram} />{/snippet}</Field>
				</fieldset>

				<div class="rounded-sm border border-border/60 p-4">
					<Checkbox
						checked={visibility === 'approved'}
						onchange={(e) => (visibility = e.currentTarget.checked ? 'approved' : 'draft')}
					>
						<span class="block text-foreground">Make my profile public</span>
						<span class="mt-1 block text-xs text-muted-foreground">
							{#if visibility === 'approved'}
								Anyone can see your profile and published works. Untick to take it offline.
							{:else}
								Only you can see your profile right now. Add a portrait and a short bio, then tick this
								and save to go live — no extra review needed.
							{/if}
						</span>
					</Checkbox>
				</div>

				{#if error}<Alert tone="error">{error}</Alert>{/if}

				<Button type="submit" loading={saving}>{saving ? 'Saving…' : 'Save profile'}</Button>
			</form>

			{#if previewProfile}
				<div class="lg:sticky lg:top-8 lg:self-start">
					<p class="t-eyebrow mb-4 text-muted-foreground">Live preview</p>
					<ShareableProfile artist={previewProfile} works={previewWorks} framed />
				</div>
			{/if}
		</div>
	{/if}
</section>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
