<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { cn } from '../cn';
	import Button from './Button.svelte';

	type Props = {
		/** HTTP status when known. Picks the right wording; 404 and auth errors hide Retry. */
		status?: number;
		title?: string;
		message?: string;
		/** Defaults to re-running every load function rather than a hard reload. */
		onretry?: () => void | Promise<void>;
		class?: string;
	};

	let { status = 500, title, message, onretry, class: className }: Props = $props();

	const copy = $derived.by(() => {
		if (status === 404) return { eyebrow: 'Not found', title: 'We can’t find that.', retry: false };
		if (status === 401 || status === 403) return { eyebrow: 'Restricted', title: 'You don’t have access to this.', retry: false };
		if (status === 503 || status === 0) return { eyebrow: 'Unreachable', title: 'We can’t reach the archive right now.', retry: true };
		return { eyebrow: 'Something went wrong', title: 'That didn’t load.', retry: true };
	});

	let retrying = $state(false);

	async function retry() {
		retrying = true;
		try {
			await (onretry ? onretry() : invalidateAll());
		} finally {
			retrying = false;
		}
	}
</script>

<div role="alert" class={cn('rounded-sm border border-border bg-card/40 p-8 md:p-12', className)}>
	<p class="t-eyebrow text-accent">{copy.eyebrow}</p>
	<h2 class="t-h3 mt-4 max-w-lg text-foreground md:text-3xl">{title ?? copy.title}</h2>
	{#if message}<p class="t-body mt-4 max-w-xl text-muted-foreground">{message}</p>{/if}
	{#if copy.retry}
		<div class="mt-6"><Button variant="secondary" loading={retrying} onclick={retry}>Try again</Button></div>
	{/if}
</div>
