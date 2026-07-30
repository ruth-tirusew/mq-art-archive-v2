<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { authService } from '$lib/application/auth';
	import { ApiError } from '$lib/adapters/api/client';

	const token = $derived($page.url.searchParams.get('token') ?? '');

	let password = $state('');
	let confirm = $state('');
	let submitting = $state(false);
	let error = $state('');
	let done = $state(false);

	async function handleSubmit(event: Event) {
		event.preventDefault();
		error = '';
		if (!token) {
			error = 'Missing reset token.';
			return;
		}
		if (password !== confirm) {
			error = 'Passwords do not match.';
			return;
		}
		submitting = true;
		try {
			await authService.resetPassword(token, password);
			done = true;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Could not reset password.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Reset password — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-lg px-6 py-24 md:px-10">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Account</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Reset password</h1>

	{#if done}
		<p class="mt-8 text-sm text-foreground">Your password has been updated.</p>
		<button
			type="button"
			class="mt-6 font-mono text-[11px] uppercase tracking-[0.2em] text-accent underline underline-offset-4"
			onclick={() => goto('/login')}
		>
			Sign in
		</button>
	{:else}
		<p class="mt-4 text-sm leading-relaxed text-muted-foreground">
			Choose a new password (at least 8 characters).
		</p>
		<form class="mt-8 space-y-4" onsubmit={handleSubmit}>
			<label class="block">
				<span class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">New password</span>
				<input
					type="password"
					required
					minlength="8"
					autocomplete="new-password"
					bind:value={password}
					class="mt-2 w-full rounded-sm border border-border bg-card px-3 py-2.5 text-sm text-foreground outline-none transition focus:border-foreground/40"
				/>
			</label>
			<label class="block">
				<span class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">Confirm password</span>
				<input
					type="password"
					required
					minlength="8"
					autocomplete="new-password"
					bind:value={confirm}
					class="mt-2 w-full rounded-sm border border-border bg-card px-3 py-2.5 text-sm text-foreground outline-none transition focus:border-foreground/40"
				/>
			</label>
			{#if error}
				<p class="text-sm text-destructive" role="alert">{error}</p>
			{/if}
			<button
				type="submit"
				disabled={submitting || !token}
				class="w-full rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
			>
				{submitting ? 'Saving…' : 'Update password'}
			</button>
		</form>
	{/if}
</section>
