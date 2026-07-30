<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { googleLoginUrl } from '$lib/core/domain/auth';
	import { authService, postLoginPath } from '$lib/application/auth';
	import { ApiError } from '$lib/adapters/api/client';

	type Mode = 'signin' | 'signup';

	const returnTo = $derived($page.url.searchParams.get('return_to') ?? '/studio');
	const loginHref = $derived(googleLoginUrl(returnTo, $page.url.origin));

	let mode = $state<Mode>('signin');
	let email = $state('');
	let password = $state('');
	let submitting = $state(false);
	let error = $state('');

	async function handleSubmit(event: Event) {
		event.preventDefault();
		error = '';
		submitting = true;
		try {
			const user =
				mode === 'signup'
					? await authService.register(email, password)
					: await authService.login(email, password);
			await goto(postLoginPath(user, returnTo));
		} catch (err) {
			if (err instanceof ApiError) {
				if (err.status === 409) {
					error = 'An account with this email already exists.';
				} else if (err.status === 401) {
					error = 'Invalid email or password.';
				} else {
					error = err.message;
				}
			} else {
				error = 'Something went wrong. Please try again.';
			}
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{mode === 'signup' ? 'Sign up' : 'Sign in'} — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-lg px-6 py-24 md:px-10">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Account</p>
	<h1 class="mt-4 font-display text-4xl text-foreground">
		{mode === 'signup' ? 'Create account' : 'Sign in'}
	</h1>
	<p class="mt-4 text-sm leading-relaxed text-muted-foreground">
		{mode === 'signup'
			? 'Create an account with email and password, or continue with Google.'
			: 'Sign in with email and password or Google to manage your profile, submit an application, or access your studio.'}
	</p>

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
		<label class="block">
			<span class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground"
				>Password</span
			>
			<input
				type="password"
				required
				minlength="8"
				autocomplete={mode === 'signup' ? 'new-password' : 'current-password'}
				bind:value={password}
				class="mt-2 w-full rounded-sm border border-border bg-card px-3 py-2.5 text-sm text-foreground outline-none transition focus:border-foreground/40"
			/>
		</label>

		{#if mode === 'signin'}
			<p class="text-right">
				<a
					href="/forgot-password"
					class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground underline underline-offset-4 hover:text-foreground"
				>
					Forgot password?
				</a>
			</p>
		{/if}

		{#if error}
			<p class="text-sm text-destructive" role="alert">{error}</p>
		{/if}

		<button
			type="submit"
			disabled={submitting}
			class="inline-flex w-full items-center justify-center rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90 disabled:opacity-50"
		>
			{submitting ? 'Please wait…' : mode === 'signup' ? 'Create account' : 'Sign in'}
		</button>
	</form>

	<p class="mt-4 text-sm text-muted-foreground">
		{#if mode === 'signin'}
			Don't have an account?
			<button
				type="button"
				class="text-foreground underline underline-offset-4"
				onclick={() => {
					mode = 'signup';
					error = '';
				}}
			>
				Sign up
			</button>
		{:else}
			Already have an account?
			<button
				type="button"
				class="text-foreground underline underline-offset-4"
				onclick={() => {
					mode = 'signin';
					error = '';
				}}
			>
				Sign in
			</button>
		{/if}
	</p>

	<div class="my-8 flex items-center gap-4">
		<div class="h-px flex-1 bg-border"></div>
		<span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">or</span>
		<div class="h-px flex-1 bg-border"></div>
	</div>

	<a
		href={loginHref}
		class="inline-flex w-full items-center justify-center gap-2 rounded-sm border border-border bg-card px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-foreground transition hover:border-foreground/30"
	>
		Continue with Google
	</a>
</section>
