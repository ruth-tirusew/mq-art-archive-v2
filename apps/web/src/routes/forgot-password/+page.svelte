<script lang="ts">
	import { authService } from '$lib/application/auth';
	import { ApiError } from '$lib/adapters/api/client';

	let email = $state('');
	let submitting = $state(false);
	let error = $state('');
	let sent = $state(false);

	async function handleSubmit(event: Event) {
		event.preventDefault();
		error = '';
		submitting = true;
		try {
			await authService.forgotPassword(email);
			sent = true;
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Something went wrong. Please try again.';
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Forgot password — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-lg px-6 py-24 md:px-10">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Account</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">Forgot password</h1>
	<p class="mt-4 text-sm leading-relaxed text-muted-foreground">
		Enter your email and we will send a reset link if an account exists.
	</p>

	{#if sent}
		<p class="mt-8 rounded-sm border border-border/60 bg-card/40 p-4 text-sm text-foreground">
			If an account exists for that email, a reset link is on its way. Check your inbox (and spam).
		</p>
		<a href="/login" class="mt-6 inline-block font-mono text-[11px] uppercase tracking-[0.2em] text-accent underline underline-offset-4">
			Back to sign in
		</a>
	{:else}
		<form class="mt-8 space-y-4" onsubmit={handleSubmit}>
			<label class="block">
				<span class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">Email</span>
				<input
					type="email"
					required
					autocomplete="email"
					bind:value={email}
					class="mt-2 w-full rounded-sm border border-border bg-card px-3 py-2.5 text-sm text-foreground outline-none transition focus:border-foreground/40"
				/>
			</label>
			{#if error}
				<p class="text-sm text-destructive" role="alert">{error}</p>
			{/if}
			<button
				type="submit"
				disabled={submitting}
				class="w-full rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background disabled:opacity-50"
			>
				{submitting ? 'Sending…' : 'Send reset link'}
			</button>
		</form>
		<a href="/login" class="mt-6 inline-block font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground hover:text-foreground">
			Back to sign in
		</a>
	{/if}
</section>
