<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { checkSession, sessionExpired, watchSession } from '$lib/application/session';

	const loginHref = $derived(`/login?return_to=${encodeURIComponent($page.url.pathname + $page.url.search)}`);

	onMount(() => {
		const unwatch = watchSession();
		// Signing in from the other tab sets the cookie here too; notice it when they come back.
		const onVisible = () => {
			if (document.visibilityState === 'visible' && $sessionExpired) void checkSession();
		};
		document.addEventListener('visibilitychange', onVisible);
		window.addEventListener('focus', onVisible);
		return () => {
			unwatch();
			document.removeEventListener('visibilitychange', onVisible);
			window.removeEventListener('focus', onVisible);
		};
	});
</script>

{#if $sessionExpired}
	<div
		class="fixed inset-x-0 bottom-0 z-50 border-t border-accent/40 bg-background/95 px-4 py-3 shadow-[0_-12px_30px_-20px_rgba(0,0,0,0.35)] backdrop-blur sm:px-6"
		role="alert"
		data-testid="session-expired"
	>
		<div class="mx-auto flex max-w-3xl flex-wrap items-center gap-x-4 gap-y-2">
			<p class="min-w-0 flex-1 text-sm leading-relaxed text-foreground">
				<span class="font-medium">You've been signed out.</span>
				Nothing on this page is lost. Sign in again in a new tab, then come back here and save.
			</p>
			<a
				href={loginHref}
				target="_blank"
				rel="noopener"
				class="inline-flex h-9 items-center rounded-sm bg-accent px-4 text-sm font-medium text-accent-foreground transition hover:opacity-90"
			>
				Sign in again
			</a>
			<button type="button" class="text-sm text-muted-foreground underline underline-offset-4 hover:text-foreground" onclick={() => checkSession()}>
				I've signed in
			</button>
		</div>
	</div>
{/if}
