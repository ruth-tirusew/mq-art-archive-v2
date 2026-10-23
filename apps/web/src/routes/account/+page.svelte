<script lang="ts">
	import Skeleton from '$lib/components/Skeleton.svelte';
	import { Alert, Button, Checkbox, Field, Input, PageHeader, toast } from '$ui';
	import { authService, currentUser } from '$lib/application/auth';
	import type { NotificationPreferences } from '$lib/core/domain/auth';

	let displayName = $state($currentUser?.display_name ?? '');
	let avatarUrl = $state($currentUser?.avatar_url ?? '');
	let profileSaving = $state(false);
	let profileError = $state('');

	async function saveProfile(event: Event) {
		event.preventDefault();
		profileSaving = true;
		profileError = '';
		try {
			await authService.updateProfile(displayName.trim(), avatarUrl.trim());
			toast.success('Profile updated.');
		} catch (e) {
			profileError = e instanceof Error ? e.message : 'Could not update profile';
		} finally {
			profileSaving = false;
		}
	}

	let email = $state($currentUser?.email ?? '');
	let emailCurrentPassword = $state('');
	let emailSaving = $state(false);
	let emailError = $state('');

	async function saveEmail(event: Event) {
		event.preventDefault();
		emailSaving = true;
		emailError = '';
		try {
			await authService.changeEmail(email.trim(), emailCurrentPassword);
			emailCurrentPassword = '';
			toast.success('Email updated.');
		} catch (e) {
			emailError = e instanceof Error ? e.message : 'Could not update email';
		} finally {
			emailSaving = false;
		}
	}

	let currentPassword = $state('');
	let newPassword = $state('');
	let passwordSaving = $state(false);
	let passwordError = $state('');

	async function savePassword(event: Event) {
		event.preventDefault();
		passwordSaving = true;
		passwordError = '';
		try {
			await authService.changePassword(currentPassword, newPassword);
			currentPassword = '';
			newPassword = '';
			toast.success('Password updated.');
		} catch (e) {
			passwordError = e instanceof Error ? e.message : 'Could not update password';
		} finally {
			passwordSaving = false;
		}
	}

	let notifications = $state<NotificationPreferences | null>(null);
	let notificationsLoading = $state(true);
	let notificationsLoadError = $state(false);
	let notificationsSaving = $state(false);
	let notificationsError = $state('');

	function loadNotifications() {
		notificationsLoading = true;
		notificationsLoadError = false;
		authService
			.getNotifications()
			.then((prefs) => (notifications = prefs))
			.catch(() => (notificationsLoadError = true))
			.finally(() => (notificationsLoading = false));
	}

	$effect(loadNotifications);

	async function saveNotifications() {
		if (!notifications) return;
		notificationsSaving = true;
		notificationsError = '';
		try {
			notifications = await authService.updateNotifications(notifications);
			toast.success('Preferences saved.');
		} catch (e) {
			notificationsError = e instanceof Error ? e.message : 'Could not save preferences';
		} finally {
			notificationsSaving = false;
		}
	}
</script>

<svelte:head><title>Profile settings — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl space-y-14 px-6 py-14 md:px-10 md:py-20">
	<PageHeader eyebrow="Account">Profile <em class="italic">settings</em></PageHeader>

	<form class="space-y-5" onsubmit={saveProfile}>
		<h2 class="t-h3">Public profile</h2>
		<Field label="Display name">
			{#snippet children({ id })}<Input {id} bind:value={displayName} placeholder="Your name" autocomplete="name" />{/snippet}
		</Field>
		<Field label="Avatar URL" hint="A link to an image. Square images work best.">
			{#snippet children({ id, describedby })}
				<Input {id} aria-describedby={describedby} type="url" bind:value={avatarUrl} placeholder="https://…" />
			{/snippet}
		</Field>
		{#if profileError}<Alert tone="error">{profileError}</Alert>{/if}
		<Button type="submit" loading={profileSaving}>{profileSaving ? 'Saving…' : 'Save profile'}</Button>
	</form>

	<form class="space-y-5 border-t border-border/60 pt-10" onsubmit={saveEmail}>
		<h2 class="t-h3">Email address</h2>
		<Field label="Email" required>
			{#snippet children({ id, required })}<Input {id} {required} type="email" autocomplete="email" bind:value={email} />{/snippet}
		</Field>
		<Field label="Confirm with your password" required>
			{#snippet children({ id, required })}
				<Input {id} {required} type="password" autocomplete="current-password" bind:value={emailCurrentPassword} />
			{/snippet}
		</Field>
		{#if emailError}<Alert tone="error">{emailError}</Alert>{/if}
		<Button type="submit" loading={emailSaving}>{emailSaving ? 'Saving…' : 'Update email'}</Button>
	</form>

	<div class="space-y-5 border-t border-border/60 pt-10">
		<h2 class="t-h3">Password</h2>
		{#if $currentUser?.has_password}
			<form class="space-y-5" onsubmit={savePassword}>
				<Field label="Current password" required>
					{#snippet children({ id, required })}
						<Input {id} {required} type="password" autocomplete="current-password" bind:value={currentPassword} />
					{/snippet}
				</Field>
				<Field label="New password" required hint="At least 8 characters.">
					{#snippet children({ id, describedby, required })}
						<Input {id} {required} aria-describedby={describedby} type="password" minlength={8} autocomplete="new-password" bind:value={newPassword} />
					{/snippet}
				</Field>
				{#if passwordError}<Alert tone="error">{passwordError}</Alert>{/if}
				<Button type="submit" loading={passwordSaving}>{passwordSaving ? 'Saving…' : 'Update password'}</Button>
			</form>
		{:else}
			<p class="t-body text-muted-foreground">
				You sign in with Google, so there is no password on this account to change.
			</p>
		{/if}
	</div>

	<div class="space-y-5 border-t border-border/60 pt-10">
		<h2 class="t-h3">Notifications</h2>
		{#if notificationsLoading}
			<div class="space-y-3" role="status" aria-live="polite">
				<span class="sr-only">Loading notification settings</span>
				<Skeleton class="h-5 w-3/4" />
				<Skeleton class="h-5 w-2/3" />
			</div>
		{:else if notificationsLoadError}
			<Alert tone="error" title="Couldn’t load your preferences">
				<div class="flex flex-wrap items-center gap-4">
					<span>Your saved settings are unchanged.</span>
					<Button variant="secondary" size="sm" onclick={loadNotifications}>Try again</Button>
				</div>
			</Alert>
		{:else if notifications}
			<Checkbox bind:checked={notifications.newsletter_enabled}>
				Weekly digest — events, art, and articles by email and Telegram
			</Checkbox>
			<Checkbox bind:checked={notifications.email_on_new_application}>Email me about my application status</Checkbox>
			{#if notificationsError}<Alert tone="error">{notificationsError}</Alert>{/if}
			<Button loading={notificationsSaving} onclick={saveNotifications}>
				{notificationsSaving ? 'Saving…' : 'Save preferences'}
			</Button>
		{/if}
	</div>
</section>
