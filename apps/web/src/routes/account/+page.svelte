<script lang="ts">
	import { authService, currentUser } from '$lib/application/auth';
	import type { NotificationPreferences } from '$lib/core/domain/auth';

	let displayName = $state($currentUser?.display_name ?? '');
	let avatarUrl = $state($currentUser?.avatar_url ?? '');
	let profileSaving = $state(false);
	let profileMessage = $state('');
	let profileError = $state('');

	async function saveProfile(event: Event) {
		event.preventDefault();
		profileSaving = true;
		profileMessage = '';
		profileError = '';
		try {
			await authService.updateProfile(displayName.trim(), avatarUrl.trim());
			profileMessage = 'Profile updated.';
		} catch (e) {
			profileError = e instanceof Error ? e.message : 'Could not update profile';
		} finally {
			profileSaving = false;
		}
	}

	let email = $state($currentUser?.email ?? '');
	let emailCurrentPassword = $state('');
	let emailSaving = $state(false);
	let emailMessage = $state('');
	let emailError = $state('');

	async function saveEmail(event: Event) {
		event.preventDefault();
		emailSaving = true;
		emailMessage = '';
		emailError = '';
		try {
			await authService.changeEmail(email.trim(), emailCurrentPassword);
			emailCurrentPassword = '';
			emailMessage = 'Email updated.';
		} catch (e) {
			emailError = e instanceof Error ? e.message : 'Could not update email';
		} finally {
			emailSaving = false;
		}
	}

	let currentPassword = $state('');
	let newPassword = $state('');
	let passwordSaving = $state(false);
	let passwordMessage = $state('');
	let passwordError = $state('');

	async function savePassword(event: Event) {
		event.preventDefault();
		passwordSaving = true;
		passwordMessage = '';
		passwordError = '';
		try {
			await authService.changePassword(currentPassword, newPassword);
			currentPassword = '';
			newPassword = '';
			passwordMessage = 'Password updated.';
		} catch (e) {
			passwordError = e instanceof Error ? e.message : 'Could not update password';
		} finally {
			passwordSaving = false;
		}
	}

	let notifications = $state<NotificationPreferences | null>(null);
	let notificationsLoading = $state(true);
	let notificationsSaving = $state(false);
	let notificationsMessage = $state('');
	let notificationsError = $state('');

	$effect(() => {
		authService
			.getNotifications()
			.then((prefs) => (notifications = prefs))
			.catch(() => {})
			.finally(() => (notificationsLoading = false));
	});

	async function saveNotifications() {
		if (!notifications) return;
		notificationsSaving = true;
		notificationsMessage = '';
		notificationsError = '';
		try {
			notifications = await authService.updateNotifications(notifications);
			notificationsMessage = 'Preferences saved.';
		} catch (e) {
			notificationsError = e instanceof Error ? e.message : 'Could not save preferences';
		} finally {
			notificationsSaving = false;
		}
	}
</script>

<svelte:head><title>Profile settings — Artiv</title></svelte:head>

<section class="mx-auto max-w-3xl space-y-14 px-6 py-14 md:px-10 md:py-20">
	<div>
		<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Account</p>
		<h1 class="mt-4 font-display text-4xl text-foreground">Profile settings</h1>
	</div>

	<form class="space-y-4" onsubmit={saveProfile}>
		<h2 class="font-display text-xl text-foreground">Public profile</h2>
		<div>
			<label for="display-name" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				Display name
			</label>
			<input id="display-name" class="field" bind:value={displayName} placeholder="Your name" />
		</div>
		<div>
			<label for="avatar-url" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				Avatar URL
			</label>
			<input id="avatar-url" class="field" bind:value={avatarUrl} placeholder="https://…" />
		</div>
		{#if profileMessage}<p class="text-sm text-accent">{profileMessage}</p>{/if}
		{#if profileError}<p class="text-sm text-destructive" role="alert">{profileError}</p>{/if}
		<button
			type="submit"
			disabled={profileSaving}
			class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
		>
			{profileSaving ? 'Saving…' : 'Save profile'}
		</button>
	</form>

	<form class="space-y-4 border-t border-border/60 pt-10" onsubmit={saveEmail}>
		<h2 class="font-display text-xl text-foreground">Email address</h2>
		<div>
			<label for="email" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
				Email
			</label>
			<input id="email" type="email" class="field" bind:value={email} required />
		</div>
		<div>
			<label
				for="email-current-password"
				class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
			>
				Confirm with your password
			</label>
			<input
				id="email-current-password"
				type="password"
				class="field"
				bind:value={emailCurrentPassword}
				required
			/>
		</div>
		{#if emailMessage}<p class="text-sm text-accent">{emailMessage}</p>{/if}
		{#if emailError}<p class="text-sm text-destructive" role="alert">{emailError}</p>{/if}
		<button
			type="submit"
			disabled={emailSaving}
			class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
		>
			{emailSaving ? 'Saving…' : 'Update email'}
		</button>
	</form>

	{#if $currentUser?.has_password}
		<form class="space-y-4 border-t border-border/60 pt-10" onsubmit={savePassword}>
			<h2 class="font-display text-xl text-foreground">Password</h2>
			<div>
				<label
					for="current-password"
					class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground"
				>
					Current password
				</label>
				<input id="current-password" type="password" class="field" bind:value={currentPassword} required />
			</div>
			<div>
				<label for="new-password" class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
					New password
				</label>
				<input id="new-password" type="password" minlength="8" class="field" bind:value={newPassword} required />
			</div>
			{#if passwordMessage}<p class="text-sm text-accent">{passwordMessage}</p>{/if}
			{#if passwordError}<p class="text-sm text-destructive" role="alert">{passwordError}</p>{/if}
			<button
				type="submit"
				disabled={passwordSaving}
				class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
			>
				{passwordSaving ? 'Saving…' : 'Update password'}
			</button>
		</form>
	{/if}

	<div class="space-y-4 border-t border-border/60 pt-10">
		<h2 class="font-display text-xl text-foreground">Notifications</h2>
		{#if notificationsLoading}
			<p class="text-sm text-muted-foreground">Loading…</p>
		{:else if notifications}
			<label class="flex items-center gap-3 text-sm text-foreground">
				<input type="checkbox" bind:checked={notifications.newsletter_enabled} />
				Weekly digest — events, art, and articles by email and Telegram
			</label>
			<label class="flex items-center gap-3 text-sm text-foreground">
				<input type="checkbox" bind:checked={notifications.email_on_new_application} />
				Email me about my application status
			</label>
			{#if notificationsMessage}<p class="text-sm text-accent">{notificationsMessage}</p>{/if}
			{#if notificationsError}<p class="text-sm text-destructive" role="alert">{notificationsError}</p>{/if}
			<button
				type="button"
				disabled={notificationsSaving}
				onclick={saveNotifications}
				class="rounded-sm bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
			>
				{notificationsSaving ? 'Saving…' : 'Save preferences'}
			</button>
		{/if}
	</div>
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
