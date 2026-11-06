<script lang="ts">
	import { page } from '$app/stores';
	import { authService } from '$lib/application/auth';
	import { inviteService } from '$lib/application/invites';
	import { ApiError } from '$lib/adapters/api/client';
	import type { InvitePreview } from '$lib/core/domain/invite';
	import { Alert, Button, Field, Input, PageHeader } from '$ui';

	type Props = {
		invite: InvitePreview;
		token: string;
		/** Called once the account exists and the session is active. */
		onregistered: () => void;
	};

	let { invite, token, onregistered }: Props = $props();

	let password = $state('');
	let submitting = $state(false);
	let error = $state('');
	let accountExists = $state(false);
	let attempted = $state(false);

	const here = $derived($page.url.pathname + $page.url.search);
	const signInHref = $derived(
		`/login?intent=artist&email=${encodeURIComponent(invite.email)}&return_to=${encodeURIComponent(here)}`
	);
	const googleHref = $derived(authService.googleLoginUrl(here, $page.url.origin));
	const passwordError = $derived(attempted && password.length < 8 ? 'Use at least 8 characters.' : undefined);
	const firstName = $derived(invite.display_name?.trim().split(/\s+/)[0] ?? '');

	async function create(event: Event) {
		event.preventDefault();
		attempted = true;
		error = '';
		accountExists = false;
		if (password.length < 8) return;
		submitting = true;
		try {
			await inviteService.register(token, password);
			await authService.load(); // picks up the new session
			onregistered();
		} catch (e) {
			if (e instanceof ApiError && e.status === 409) {
				// Either this address already has an account, or the invite was just used or cancelled.
				accountExists = true;
			} else if (e instanceof ApiError && e.status === 429) {
				error = 'Too many attempts. Please wait a minute and try again.';
			} else if (e instanceof ApiError && e.status === 400) {
				error = e.message.replace(/^validation failed:\s*/i, '');
			} else {
				error = 'We couldn’t create your account just now. Please try again.';
			}
		} finally {
			submitting = false;
		}
	}
</script>

<section class="mx-auto max-w-xl px-6 py-16 md:px-10 md:py-24" data-testid="invite-landing">
	<PageHeader
		eyebrow="You’re invited"
		lead="The Artiv team invited you to join as an artist. Create a password to get started. It takes about a minute, and you’ll answer a few questions about your work next."
	>
		{#if firstName}Welcome, <em class="italic">{firstName}</em>{:else}Welcome to <em class="italic">Artiv</em>{/if}
	</PageHeader>

	<form class="mt-10 space-y-6" novalidate onsubmit={create}>
		<div class="rounded-sm border border-border/70 bg-card/30 p-4 text-sm">
			<p class="t-eyebrow text-muted-foreground">Your account</p>
			<p class="mt-2 break-all text-foreground">{invite.email}</p>
			<p class="mt-1 text-xs text-muted-foreground">The address your invitation was sent to.</p>
		</div>

		<Field label="Choose a password" required error={passwordError} hint="At least 8 characters.">
			{#snippet children({ id, describedby, invalid, required })}
				<Input
					{id}
					{required}
					{invalid}
					type="password"
					aria-describedby={describedby}
					autocomplete="new-password"
					bind:value={password}
				/>
			{/snippet}
		</Field>

		{#if accountExists}
			<Alert tone="warning">
				<p>
					<strong>{invite.email}</strong> already has an Artiv account (or this invitation has just been used).
				</p>
				<p class="mt-2"><a href={signInHref} class="underline underline-offset-4">Sign in to continue</a></p>
			</Alert>
		{/if}
		{#if error}<Alert tone="error">{error}</Alert>{/if}

		<Button type="submit" loading={submitting}>{submitting ? 'Creating your account…' : 'Create account and continue'}</Button>

		<p class="text-sm text-muted-foreground">
			Already have an account?
			<a href={signInHref} class="text-accent underline underline-offset-4">Sign in</a>
			<span aria-hidden="true"> · </span>
			<a href={googleHref} class="text-accent underline underline-offset-4">Continue with Google</a>
		</p>
	</form>
</section>
