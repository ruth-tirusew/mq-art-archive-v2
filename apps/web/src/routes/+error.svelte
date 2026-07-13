<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="mx-auto max-w-2xl px-6 py-32 text-center">
	<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">Not found</p>
	<h1 class="mt-4 font-display text-5xl text-foreground">{$page.status}</h1>
	<p class="mt-4 text-muted-foreground">{$page.error?.message}</p>
	<a
		href="/"
		class="mt-8 inline-block font-mono text-[11px] uppercase tracking-[0.2em] text-foreground underline decoration-accent underline-offset-8"
	>
		Back home
	</a>
</div>
